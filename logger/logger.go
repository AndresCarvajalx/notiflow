package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func Init() {
	os.Mkdir("logs", os.ModePerm)

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"logs/notiflow.log", "stdout"}
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()

	if err != nil {
		panic(err)
	}

	L = logger
}
