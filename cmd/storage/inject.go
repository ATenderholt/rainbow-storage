//go:build wireinject
// +build wireinject

package main

import (
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/ATenderholt/rainbow-storage/internal/http"
	"github.com/ATenderholt/rainbow-storage/internal/service"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"github.com/google/wire"
)

var api = wire.NewSet(
	http.NewChiMux,
	http.NewMinioHandler,
)

func mapConfig(cfg *settings.Config) service.Config {
	return cfg
}

var services = wire.NewSet(
	service.NewNotificationService,
	service.NewConfigurationService,
	wire.Bind(new(http.NotificationService), new(*service.NotificationService)),
	wire.Bind(new(http.ConfigurationService), new(*service.ConfigurationService)),
	mapConfig,
)

func InjectApp(cfg *settings.Config) (App, error) {
	wire.Build(
		NewApp,
		NewLambdaInvoker,
		wire.Bind(new(domain.CloudFunctionInvoker), new(*LambdaInvoker)),
		api,
		services,
		dockerlib.NewDockerController,
	)
	return App{}, nil
}
