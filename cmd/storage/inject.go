//go:build wireinject
// +build wireinject

package main

import (
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow-storage/internal/http"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/google/wire"
)

var api = wire.NewSet(
	http.NewChiMux,
	http.NewMinioHandler,
)

func InjectApp(cfg *settings.Config) (App, error) {
	wire.Build(
		NewApp,
		api,
		dockerlib.NewDockerController,
	)
	return App{}, nil
}
