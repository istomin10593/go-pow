package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger.
func New(debug string) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	level := zapcore.InfoLevel

	if debug == "true" {
		level = zapcore.DebugLevel
	}

	config.Level = zap.NewAtomicLevelAt(level)

	logger, _ := config.Build()

	return logger
}
