package main

import (
	"context"
	"flag"
	"go-pow/client/internal/client"
	"go-pow/client/pkg/config"
	"go-pow/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

const (
	defaultConfigFile = "conf.yaml"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	// Get logger instance.
	debug, _ := os.LookupEnv("DEBUG")

	log := logger.New(debug)
	defer func() {
		done()
		_ = log.Sync()

		if r := recover(); r != nil {
			log.Fatal("application panic", zap.Any("panic", r))
		}
	}()

	// Parse config.
	confFlag := flag.String("conf", "", "config yaml file")
	flag.Parse()

	confString := *confFlag
	if confString == "" {
		confString = defaultConfigFile
	}

	// Get application cfg.
	cfg, err := config.Parse(confString)
	if err != nil {
		log.Panic("failed to parse config", zap.Error(err))
	}

	// Channel to notify when client is shut down.
	shutdownComplete := make(chan struct{})
	defer close(shutdownComplete)

	// Run client.
	go func() {
		client.New(log, cfg).Run(ctx)

		// Notify client shutting down.
		log.Info("client shutting down...")

		shutdownComplete <- struct{}{}
	}()

	<-shutdownComplete

	log.Info("successful shutdown")
}
