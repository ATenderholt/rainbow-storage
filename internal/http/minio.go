package http

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v2"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

type NotificationService interface {
	GetConfigurationPath(bucket string) string
	ProcessEvent(bucket string, event domain.NotificationEvent) error
	Save(bucket string, config domain.NotificationConfiguration) (string, error)
}

type ResponseWriter struct {
	http.ResponseWriter
	Code *int
}

func (w ResponseWriter) WriteHeader(code int) {
	*w.Code = code
	w.ResponseWriter.WriteHeader(code)
}

type MinioHandler struct {
	cfg     *settings.Config
	service NotificationService
}

func NewMinioHandler(cfg *settings.Config, service NotificationService) MinioHandler {
	return MinioHandler{
		cfg:     cfg,
		service: service,
	}
}

func (h MinioHandler) GetNotifications(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, request *http.Request) {
		if !request.URL.Query().Has("notification") {
			next.ServeHTTP(w, request)
			return
		}

		bucket := chi.URLParam(request, "bucket")
		logger.Infof("Loading NotificationConfiguration for bucket %s", bucket)

		path := h.service.GetConfigurationPath(bucket)
		file, err := os.Open(path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			logger.Warnf("File %s does not exist: %v", path, err)
			http.NotFound(w, request)
			return
		case err != nil:
			logger.Errorf("Error when opening file %s: %v", path, err)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		decoder := yaml.NewDecoder(file)

		var notification domain.NotificationConfiguration
		err = decoder.Decode(&notification)
		if err != nil {
			logger.Errorf("Unable to decode NotificationConfiguration for bucket %s: %v", bucket, err)
			http.Error(w, "Unable to decode NotificationConfiguration", http.StatusInternalServerError)
			return
		}

		encoder := xml.NewEncoder(w)
		err = encoder.Encode(notification)
		if err != nil {
			logger.Errorf("Unable to encode NotificationConfiguration for bucket %s: %v", bucket, err)
			http.Error(w, "Unable to encode NotificationConfiguration", http.StatusInternalServerError)
			return
		}
	}

	return http.HandlerFunc(f)
}

func (h MinioHandler) PutNotifications(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, request *http.Request) {
		if !request.URL.Query().Has("notification") {
			next.ServeHTTP(w, request)
			return
		}

		bucket := chi.URLParam(request, "bucket")
		logger.Infof("Saving NotificationConfiguration for bucket %s", bucket)

		payload, _ := io.ReadAll(request.Body)
		request.Body.Close()

		var notification domain.NotificationConfiguration
		err := xml.Unmarshal(payload, &notification)
		if err != nil {
			msg := fmt.Sprintf("unable to unmarshall notification %s: %v", string(payload), err)
			logger.Error(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		logger.Infof("Received Notification %+v for URL %s", notification, request.URL.Path)

		if len(notification.CloudFunctionConfigurations) == 0 {
			logger.Infof("No configuration found fo raw payload: %s", string(payload))
			logger.Infof("Query params: %v", request.URL.RawQuery)
			http.Error(w, "No configuration functions", http.StatusBadRequest)
			return
		}

		_, err = h.service.Save(bucket, notification)
		if err != nil {
			logger.Errorf("Unable to save notification for bucket %s", bucket)
			http.Error(w, "Unable to save notification for bucket "+bucket, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(f)
}

func (h MinioHandler) SendNotifications(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, request *http.Request) {
		wrapped := ResponseWriter{
			ResponseWriter: w,
			Code:           new(int),
		}

		next.ServeHTTP(wrapped, request)

		bucket := chi.URLParam(request, "bucket")
		key := chi.URLParam(request, "*")

		if key == "" {
			return
		}

		// Example finished uploads:
		// PUT /myaws-files/AWSLogs/small.log
		// POST /myaws-files/AWSLogs/test.log?uploadId=1bc7323b-ad52-4ca5-9606-e1e22c38cbbd

		// Example start of uploading parts
		// POST http://localhost:9000/myaws-files/AWSLogs/test.log?uploads
		// don't send notification yet
		if request.Method == http.MethodPost && request.URL.Query().Has("uploads") {
			return
		}
		
		// Example part upload:
		// PUT http://localhost:9000/myaws-files/AWSLogs/test.log?uploadId=956d38ed-a2ef-4149-9382-3f4a819e503d&partNumber=2
		// don't send notifications for parts yet
		if request.Method == http.MethodPut && request.URL.Query().Has("uploadId") {
			return
		}

		if *wrapped.Code != http.StatusOK {
			logger.Warnf("Multipart upload for key %s in bucket %s did not finish correctly", key, bucket)
			return
		}

		logger.Infof("Completed upload for key %s in bucket %s", key, bucket)
		event := domain.NotificationEvent{
			Key:   key,
			Event: domain.ObjectCreatedEvent,
		}

		err := h.service.ProcessEvent(bucket, event)
		if err != nil {
			logger.Warnf("Unable to send event for key %s in bucket %s: %v", key, bucket, err)
		}
	}

	return http.HandlerFunc(f)
}

func (h MinioHandler) Proxy(w http.ResponseWriter, request *http.Request) {
	url := h.cfg.MinioUrl() + request.URL.Path + "?" + request.URL.RawQuery
	logger.Infof("Forwarding to %s", url)

	payload, err := io.ReadAll(request.Body)
	request.Body.Close()

	reader := bytes.NewReader(payload)
	proxyReq, _ := http.NewRequest(request.Method, url, reader)

	credentials := aws.Credentials{AccessKeyID: "minio", SecretAccessKey: "miniosecret"}

	signer := v4.NewSigner()
	err = signer.SignHTTP(request.Context(), credentials, proxyReq,
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"s3", h.cfg.Region, time.Now())

	if err != nil {
		msg := fmt.Sprintf("Unable to sign request to Minio: %v", err)
		logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		msg := fmt.Sprintf("Unable to proxy to Minio: %v", err)
		logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}

	}

	var response strings.Builder
	body := io.TeeReader(resp.Body, &response)

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, body)

	if resp.StatusCode != 200 {
		logger.Infof("Response (%d): %s", resp.StatusCode, response.String())
		logger.Infof("Request payload: %s", payload)
	}

	return
}
