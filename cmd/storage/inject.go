//go:build wireinject
// +build wireinject

package main

import (
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/google/wire"
)

func NewApp(cfg *settings.Config) App {
	return App{cfg}
}

func InjectApp(cfg *settings.Config) (App, error) {
	wire.Build(
		NewApp,
	)
	return App{}, nil
}
