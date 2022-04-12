package main

import (
	"context"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow-storage/internal/service"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"path/filepath"
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

	container := dockerlib.Container{
		Name:  "s3",
		Image: app.cfg.Image,
		Mounts: []mount.Mount{
			{
				Source: basePath,
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
}

func NewLambdaInvoker() *LambdaInvoker {
	return &LambdaInvoker{}
}

func (i LambdaInvoker) Invoke(bucket string) func(interface{}) {
	return func(i interface{}) {
		logger.Infof("Processing %+v for bucket %s", i, bucket)
	}
}
