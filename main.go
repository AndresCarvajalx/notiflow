package main

import (
	"flag"
	"fmt"

	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/dashboard"
	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/license"
	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/notifier"
	"github.com/AndresCarvajalx/notiflow/scheduler"
	"github.com/AndresCarvajalx/notiflow/utils"
	"github.com/AndresCarvajalx/notiflow/validator"
	"github.com/AndresCarvajalx/notiflow/wmeow"
	"github.com/joho/godotenv"
)

func main() {
	runNotifier := flag.Bool("run", false, "Ejecutar el worker de notificaciones")
	useWhatsmeow := flag.Bool("whatsmeow", false, "Usar whatsmeow en vez de WhatsApp API")
	validate := flag.Bool("validate", false, "Validar configuración y archivo Excel")
	installSchedule := flag.Bool("install-schedule", false, "Instalar tarea programada diaria (Task Scheduler)")
	removeSchedule := flag.Bool("remove-schedule", false, "Eliminar tarea programada")

	flag.Usage = func() {
		fmt.Println(`Notiflow - Sistema de notificaciones WhatsApp

USO:
  notiflow [flags]

FLAGS:
  --help                    Muestra esta ayuda

  --run                     Ejecuta el worker de notificaciones una sola vez
                            (usado por Task Scheduler automaticamente)

  --whatsmeow               Usa WhatsApp personal.
                            Solo tiene efecto combinado con --run o --install-schedule.

  --validate                Valida que el Excel y la configuración sean correctos
                            sin enviar notificaciones.

  --install-schedule        Instala una tarea programada diaria a las 09:00.
                            Opcional: combinar con --whatsmeow.

  --remove-schedule         Elimina la tarea programada instalada.

EJEMPLOS:
  notiflow                          Inicia el dashboard web (http://localhost:<PUERTO ESPECIFICADO>)
  notiflow --run                    Ejecuta notificaciones via WhatsApp API
  notiflow --run --whatsmeow        Ejecuta notificaciones via WhatsApp Web
  notiflow --validate               Valida Excel y configuración
  notiflow --install-schedule       Instala tarea diaria con API de Meta
  notiflow --install-schedule --whatsmeow   Instala tarea diaria con WhatsApp Web
  notiflow --remove-schedule        Desinstala la tarea programada`)
	}

	flag.Parse()

	godotenv.Load()
	logger.Init()

	if *installSchedule {
		scheduler.Install(*useWhatsmeow)
		return
	}
	if *removeSchedule {
		scheduler.Remove()
		return
	}

	serial, err := license.GetWindowSerialNumber()
	if err != nil {
		utils.ShowDialog("Error obteniendo serial", "Ocurrio un error obteniendo el serial, error: "+err.Error())
		logger.L.Sugar().Fatal(err.Error())
	}
	logger.L.Sugar().Infof("Numero de serial:%s", serial)

	active, err := license.ValidateLicense(serial)
	if err != nil || !active {
		utils.ShowDialog("Error", err.Error())
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

	if *validate {
		result := validator.Run(config.Get())
		fmt.Print(result.String())
		if !result.Valid {
			logger.L.Sugar().Fatal("Validación fallida")
		}
		return
	}

	if *runNotifier {
		if *useWhatsmeow {
			if err := wmeow.Run(); err != nil {
				utils.ShowDialog("Error notifier wmeow", err.Error())
				logger.L.Sugar().Fatalf("Error running whatsmeow: %v", err.Error())
			}
		} else {
			if err := notifier.Run(); err != nil {
				utils.ShowDialog("Error notifier", err.Error())
				logger.L.Sugar().Fatalf("Error running notifier: %v", err.Error())
			}
		}
		return
	}

	port := config.Get().Server.Port
	dashboard.Init(db)
	dashboard.StartServer(port)
}
