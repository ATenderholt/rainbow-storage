package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow-storage/internal/logging"
	"github.com/ATenderholt/rainbow-storage/internal/settings"
	"go.uber.org/zap"
	"os"
	"os/signal"
)

var logger *zap.SugaredLogger

func init() {
	logger = logging.NewLogger()
}

func main() {
	cfg, output, err := settings.FromFlags(os.Args[0], os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(output)
		os.Exit(2)
	} else if err != nil {
		fmt.Println("got error:", err)
		fmt.Println("output:\n", output)
		os.Exit(1)
	}

	mainCtx := context.Background()

	dockerlib.SetLogger(logging.NewLogger().Desugar().Named("dockerlib"))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(mainCtx)
	go func() {
		s := <-c
		logger.Infof("Received signal %v", s)
		cancel()
	}()

	if err := start(ctx, cfg); err != nil {
		logger.Errorf("Failed to start: %v", err)
	}
}

func start(ctx context.Context, config *settings.Config) error {
	logger.Info("Starting up ...")

	app, err := InjectApp(config)
	if err != nil {
		logger.Errorf("Unable to initialize application: %v", err)
		return err
	}

	err = app.Start()
	if err != nil {
		logger.Errorf("Unable to start application: %v", err)
		return err
	}

	<-ctx.Done()

	logger.Info("Shutting down ...")
	err = app.Shutdown()
	if err != nil {
		logger.Error("Error when shutting down app")
	}

	return nil
}
