package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DoNotLogToFile = ""
)

func CreateLogger(l zapcore.Level, filePath string, withStdoutLogger bool) *zap.Logger {
	stdout := zapcore.AddSync(os.Stdout)

	level := zap.NewAtomicLevelAt(l)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	loggerStreams := []zapcore.Core{}

	if withStdoutLogger {
		consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
		loggerStreams = append(loggerStreams, zapcore.NewCore(consoleEncoder, stdout, level))
	}

	if len(filePath) > 0 {
		file := zapcore.AddSync(&lumberjack.Logger{
			Filename: filePath,
			MaxSize:  300,
		})

		fileEncoder := zapcore.NewConsoleEncoder(productionCfg)
		loggerStreams = append(loggerStreams, zapcore.NewCore(fileEncoder, file, level))
	}

	core := zapcore.NewTee(
		loggerStreams...,
	)

	return zap.New(core)
}
