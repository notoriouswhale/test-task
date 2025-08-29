package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(environment string, level zapcore.Level) *zap.Logger {
	var config zap.Config

	switch environment {
	case "production":
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	logger, err := config.Build()
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	return logger
}

func NewSugaredLogger(environment string, level zapcore.Level) *zap.SugaredLogger {
	return NewLogger(environment, level).Sugar()
}
