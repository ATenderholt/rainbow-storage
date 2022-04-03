package main

import (
	"context"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"time"
)

type App struct {
	cfg *settings.Config
}

func (app App) Start() (err error) {
	return nil
}

func (app App) Shutdown() error {
	_, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return nil
}
