package main

import (
	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/license"
	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/notifier"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	logger.Init()
	_, err := config.LoadConfig("config.yml")
	if err != nil {
		logger.L.Sugar().Fatalf("Error loading config: %v", err.Error())
	}

	serialNumber, err := license.GetWindowSerialNumber()
	if err != nil {
		logger.L.Sugar().Fatalf("Error getting window serial number: %v", err.Error())
	}
	logger.L.Sugar().Infof("Serial number: %s", serialNumber)

	active, err := license.ValidateLicense(serialNumber)
	if err != nil {
		logger.L.Sugar().Fatalf("Error validating license: %v", err.Error())
	}
	if !active {
		logger.L.Sugar().Fatalf("License is not active")
	}

	logger.L.Sugar().Infof("License is active, starting application")

	db := database.GetConnection()
	database.Init(db)
	defer db.Close()
	defer logger.L.Sync()

	if err := notifier.Run(); err != nil {
		logger.L.Sugar().Fatalf("Error running notifier: %v", err.Error())
	}
}
