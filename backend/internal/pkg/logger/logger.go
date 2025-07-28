package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func SetupLogger(env string, logFilePath string) *zap.Logger {
	var config zap.Config
	var encoderCfg zapcore.EncoderConfig

	if env == "prod" {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}

	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var outputPaths []string
	if env == "prod" {
		outputPaths = []string{"stderr", logFilePath}
	} else {
		outputPaths = []string{"stdout", logFilePath}
	}

	switch env {
	case "prod":
		config = zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths:       outputPaths,
			ErrorOutputPaths:  []string{"stderr"},
			InitialFields:     map[string]interface{}{"pid": os.Getpid()},
		}
	default:
		config = zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:       true,
			DisableCaller:     false,
			DisableStacktrace: false,
			Encoding:          "console",
			EncoderConfig:     encoderCfg,
			OutputPaths:       outputPaths,
			ErrorOutputPaths:  []string{"stderr"},
			InitialFields:     map[string]interface{}{"pid": os.Getpid()},
		}
	}

	if logFilePath != "" {
		logFileDir := filepath.Dir(logFilePath)
		if _, err := os.Stat(logFileDir); os.IsNotExist(err) {
			err := os.MkdirAll(logFileDir, 0755)
			if err != nil {
				panic("ошибка при создании директории для логов" + err.Error())
			}
		}
	}

	logger, err := config.Build()
	if err != nil {
		panic("ошибка создания логгера" + err.Error())
	}
	return logger
}
