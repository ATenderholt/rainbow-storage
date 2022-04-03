package main

import (
	"context"
	"fmt"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type App struct {
	cfg *settings.Config
	srv *http.Server
}

func NewApp(cfg *settings.Config, mux *chi.Mux) App {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.BasePort),
		Handler: mux,
	}

	return App{
		cfg: cfg,
		srv: srv,
	}
}

func (app App) Start() error {
	errors := make(chan error)

	go app.StartHttp(errors)

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

func (app App) Shutdown() error {
	_, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return nil
}
