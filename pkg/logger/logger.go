package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func InitLogger() {
	dev := false
	encoderCfg := zap.NewProductionEncoderConfig()
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	if err := os.MkdirAll("logs", 0o755); err != nil {
		log.Printf("Warning: failed to create logs directory: %v", err)
	}

	config := zap.Config{
		Level:             level,
		Development:       dev,
		DisableStacktrace: true,
		DisableCaller:     false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths:       make([]string, 0, 2),
		ErrorOutputPaths:  make([]string, 0, 2),
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}
	config.OutputPaths = append(config.OutputPaths, "logs/log.txt", "stdout")
	config.ErrorOutputPaths = append(config.ErrorOutputPaths, "logs/error.txt", "stderr")

	baseLogger, err := config.Build()
	if err != nil {
		log.Fatal("Error building zap logger")
	}

	Logger = baseLogger.Sugar()
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
