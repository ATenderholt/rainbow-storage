package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/ATenderholt/rainbow-storage/internal/service"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type App struct {
	cfg           *settings.Config
	docker        *dockerlib.DockerController
	notifyService *service.NotificationService
	srv           *http.Server
}

func NewApp(cfg *settings.Config, docker *dockerlib.DockerController, notifyService *service.NotificationService, mux *chi.Mux) App {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.BasePort),
		Handler: mux,
	}

	return App{
		cfg:           cfg,
		docker:        docker,
		notifyService: notifyService,
		srv:           srv,
	}
}

func (app App) Start() error {
	errors := make(chan error, 5)

	go app.StartHttp(errors)

	app.StartDocker(errors)
	app.StartNotifications(errors)

	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}

func (app *App) StartHttp(errors chan error) {
	logger.Infof("Starting HTTP server on port %d", app.cfg.BasePort)
	err := app.srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		logger.Errorf("Problem starting HTTP server: %v", err)
		errors <- err
	}
}

func (app *App) StartDocker(errors chan error) {
	logger.Infof("Starting background storage container")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err := app.docker.EnsureImage(ctx, app.cfg.Image)
	if err != nil {
		e := fmt.Errorf("unable to ensure that image %s exists: %v", app.cfg.Image, err)
		logger.Error(e)
		errors <- e
		return
	}

	basePath := filepath.Join(app.cfg.DataPath(), "buckets")
	logger.Infof("Creating directory if necessary %s ...", basePath)

	err = os.MkdirAll(basePath, 0755)
	if err != nil {
		e := fmt.Errorf("unable to create directory %s: %v", basePath, err)
		logger.Error(e)
		errors <- e
		return
	}

	var mountPath string
	if app.cfg.IsLocal {
		mountPath = basePath
	} else {
		containerName := os.Getenv("NAME")
		logger.Infof("Getting source for mount %s in container %s", app.cfg.DataPath(), containerName)
		hostPath, err := app.docker.GetContainerHostPath(ctx, containerName, app.cfg.DataPath())
		mountPath = strings.Replace(basePath, app.cfg.DataPath(), hostPath, 1)

		if err != nil {
			e := fmt.Errorf("unable to get host path for %s: %v", app.cfg.DataPath(), err)
			logger.Error(e)
			errors <- e
			return
		}
	}

	container := dockerlib.Container{
		Name:  "s3",
		Image: app.cfg.Image,
		Mounts: []mount.Mount{
			{
				Source: mountPath,
				Target: "/data",
				Type:   mount.TypeBind,
			},
		},
		Ports: map[int]int{
			9000: app.cfg.BasePort + 1,
			9001: app.cfg.BasePort + 2,
		},
		Network: app.cfg.Networks,
	}

	ready, err := app.docker.Start(ctx, &container, "Documentation: https://docs.min.io")
	if err != nil {
		e := fmt.Errorf("unable to start container: %v", err)
		logger.Error(e)
		errors <- e
		return
	}

	<-ready

	logger.Info("Background storage container is ready")

	return
}

func (app App) StartNotifications(errors chan error) {
	err := app.notifyService.LoadAll()
	if err != nil {
		errors <- err
	}
}

func (app App) Shutdown() error {
	logger.Info("Starting shutdown of application")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err := app.docker.ShutdownAll(ctx)
	if err != nil {
		logger.Error("Unable to shutdown Docker containers: %v", err)
	}

	err = app.srv.Shutdown(ctx)
	if err != nil {
		logger.Error("Unable to shutdown HTTP server: %v", err)
	}

	logger.Info("Finished shutting down application")
	return err
}

type LambdaInvoker struct {
	cfg    *settings.Config
	client *lambda.Client
}

func NewLambdaInvoker(cfg *settings.Config) *LambdaInvoker {
	return &LambdaInvoker{
		cfg:    cfg,
		client: NewLambdaClient(cfg),
	}
}

func (l LambdaInvoker) Invoke(lambdaArn string) func(interface{}) {
	return func(i interface{}) {
		value := i.(domain.NotificationEvent)

		logger.Infof("Processing %+v for lambdaArn %s", value, lambdaArn)
		parts := strings.Split(lambdaArn, ":")

		record := domain.LambdaRecord{
			EventVersion: "2.1",
			EventSource:  "aws:s3",
			AwsRegion:    l.cfg.Region,
			EventTime:    domain.JsonTime(time.Now()),
			EventName:    value.Event,
			UserIdentity: domain.LambdaUserIdentity{},
			RequestParameters: domain.LambdaRequestParameters{
				SourceIPAddress: value.SourceIp,
			},
			ResponseElements: domain.LambdaResponseElements{},
			S3: domain.S3Record{
				S3SchemaVersion: "1.0",
				ConfigurationId: "",
				Bucket: domain.S3Bucket{
					Name: value.Bucket,
					OwnerIdentity: domain.S3BucketOwnerIdentity{
						PrincipalId: "",
					},
					Arn: "arn:aws:s3:::" + value.Bucket,
				},
				Object: domain.S3Object{
					Key:       value.Key,
					Size:      value.Size,
					ETag:      "",
					Sequencer: "",
				},
			},
		}

		payload, err := json.Marshal(record)
		if err != nil {
			logger.Infof("Unable to marshal record for %+v: %v", i, err)
		}

		params := lambda.InvokeInput{
			FunctionName:   &parts[6],
			ClientContext:  nil,
			InvocationType: types.InvocationTypeEvent,
			LogType:        "",
			Payload:        payload,
			Qualifier:      nil,
		}

		_, err = l.client.Invoke(context.Background(), &params)
		if err != nil {
			return
		}
	}
}
