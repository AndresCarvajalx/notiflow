package main

import (
	"flag"

	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/dashboard"
	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/license"
	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/notifier"
	"github.com/AndresCarvajalx/notiflow/utils"
	"github.com/joho/godotenv"
)

func main() {
	runNotifier := flag.Bool("run", false, "Run notification worker")
	flag.Parse()

	godotenv.Load()

	logger.Init()

	serial, err := license.GetWindowSerialNumber()
	if err != nil {
		utils.ShowDialog("Error obteniendo serial", "Ocurrio un error obteniendo el serial, error: "+err.Error())
		logger.L.Sugar().Fatal(err.Error())
	}

	logger.L.Sugar().Infof("Numero de serial:%s", serial)

	active, err := license.ValidateLicense(serial)
	if err != nil {
		utils.ShowDialog("Error validando licencia", err.Error())
	}

	if !active {
		utils.ShowDialog("Licencia Invalida", "Su licencia expiro, por favor pongase en contacto para renovarla")
		return
	}

	db := database.GetConnection()
	database.Init(db)
	defer db.Close()
	defer logger.L.Sync()

	_, err = config.LoadConfig("config.yml")
	if err != nil {
		utils.ShowDialog("Error leyendo el archivo config.yml", "El archivo de configuracion no se pudo leer o no existe")
		logger.L.Sugar().Fatalf("Error loading config: %v", err.Error())
	}

	if *runNotifier {
		if err := notifier.Run(); err != nil {
			logger.L.Sugar().Fatalf("Error running notifier: %v", err.Error())
		}
		return
	}

	port := config.Get().Server.Port
	dashboard.Init(db)
	dashboard.StartServer(port)
}
