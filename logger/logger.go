package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func Init() {
	os.MkdirAll("./logs", 0755)
	cleanOldLogs(30)

	filename := fmt.Sprintf("./logs/notiflow-%s.log", time.Now().Format("2006-01-02"))

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	fileWriter := zapcore.AddSync(file)
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

func cleanOldLogs(maxDays int) {
	files, err := filepath.Glob("./logs/notiflow-*.log")
	if err != nil {
		return
	}
	threshold := time.Now().AddDate(0, 0, -maxDays)
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.ModTime().Before(threshold) {
			os.Remove(f)
		}
	}
}
