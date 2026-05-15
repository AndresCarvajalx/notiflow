package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L *zap.Logger

func Init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/notiflow.log",
		MaxSize:    10,
		MaxBackups: 30,
		MaxAge:     30,
		Compress:   true,
	})

	consoleWriter := zapcore.AddSync(os.Stdout)

	fileCore := zapcore.NewCore(
		encoder,
		fileWriter,
		zap.InfoLevel,
	)

	consoleCore := zapcore.NewCore(
		encoder,
		consoleWriter,
		zap.InfoLevel,
	)

	core := zapcore.NewTee(
		fileCore,
		consoleCore,
	)

	L = zap.New(core, zap.AddCaller())
}
