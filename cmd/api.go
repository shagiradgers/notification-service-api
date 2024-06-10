package main

import (
	"context"
	"log/slog"
	"notification-service-api/internal/app"
	"notification-service-api/internal/clients"
	"os"
	"os/signal"
	"syscall"

	projectConfig "notification-service-api/internal/config"
)

var logger *slog.Logger

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func main() {
	ctx := context.Background()
	config, err := projectConfig.NewConfig(projectConfig.LocalEnv)
	if err != nil {
		logger.Error("can't create config: ", err)
		return
	}

	c := clients.NewClients(ctx, config)
	if err != nil {
		logger.Error("can't create clients: ", err)
		return
	}

	a := app.New(logger, c, config.MustGetServerHost(), config.MustGetServerPort())
	go func() {
		if err = a.Run(); err != nil {
			logger.Error("can't run app: ", err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	a.Stop()
	logger.Info("Gracefully stopped")
}
