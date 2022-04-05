package http

import (
	"fmt"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"io"
	"net/http"
	"strings"
	"time"
)

type MinioHandler struct {
	cfg *settings.Config
}

func NewMinioHandler(cfg *settings.Config) MinioHandler {
	return MinioHandler{
		cfg: cfg,
	}
}

func (h MinioHandler) Proxy(w http.ResponseWriter, request *http.Request) {
	url := h.cfg.MinioUrl() + request.URL.Path
	logger.Infof("Forwarding to %s", url)

	var payload strings.Builder
	reader := io.TeeReader(request.Body, &payload)
	proxyReq, _ := http.NewRequest(request.Method, url, reader)

	credentials := aws.Credentials{AccessKeyID: "minio", SecretAccessKey: "miniosecret"}

	signer := v4.NewSigner()
	err := signer.SignHTTP(request.Context(), credentials, proxyReq,
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
		logger.Infof("Request payload: %s", payload.String())
	}

	return
}
