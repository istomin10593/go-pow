package main

import (
	"context"
	"flag"
	"fmt"
	"go-pow/server/internal/server"
	"go-pow/server/pkg/book"
	"go-pow/server/pkg/config"
	"go-pow/server/pkg/logger"
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

	// Run server.
	return server.New(log, cfg, book).Run(ctx)
}
