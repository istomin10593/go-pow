package main

import (
	"context"
	"flag"
	"fmt"
	"go-pow/pkg/logger"
	"go-pow/server/internal/handler"
	"go-pow/server/internal/server"
	"go-pow/server/pkg/book"
	"go-pow/server/pkg/cache"
	"go-pow/server/pkg/config"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

const (
	defaultConfigFile = "conf.yaml"
	defaultSourceFile = "source/quotes.txt"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	// Get logger instance.
	debug, _ := os.LookupEnv("DEBUG")

	log := logger.New(debug)

	defer func() {
		done()
		log.Sync()

		if r := recover(); r != nil {
			log.Fatal("application panic", zap.Any("panic", r))
		}
	}()

	// Run server.
	err := realMain(ctx, log)

	if err != nil {
		log.Fatal("fatal err", zap.Error(err))
	}

	log.Info("successful shutdown")
}

func realMain(ctx context.Context, log *zap.Logger) error {
	confFlag := flag.String("conf", "", "config yaml file")
	sourceFlag := flag.String("source", "", "source txt file")
	flag.Parse()

	confString := *confFlag
	if confString == "" {
		confString = defaultConfigFile
	}

	// Get application cfg.
	cfg, err := config.Parse(confString)
	if err != nil {
		log.Error("failed to parse config", zap.Error(err))

		return fmt.Errorf("failed to parse config: %w", err)
	}

	// New book source
	sourceString := *sourceFlag
	if sourceString == "" {
		sourceString = defaultSourceFile
	}

	book, err := book.New(sourceString)
	if err != nil {
		log.Error("failed to create new book", zap.Error(err))

		return fmt.Errorf("failed to create new book: %w", err)
	}

	// Get cache instance.
	cache, err := cache.New(ctx, cfg.Cache.Host, cfg.Cache.Port, cfg.Cache.Expiration)
	if err != nil {
		log.Error("failed to create new cache", zap.Error(err))

		return fmt.Errorf("failed to create new cache: %w", err)
	}

	// Create handler.
	handler := handler.New(cfg.Pow.ZeroBits, cfg.Server.Timeout, log, book, cache)

	// Run server.
	return server.New(cfg, book, handler).Run(ctx, log)
}
