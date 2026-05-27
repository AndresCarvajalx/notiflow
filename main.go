package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

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
	godotenv.Load()

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		if _, err := os.Stat(filepath.Join(exeDir, "config.yml")); err == nil {
			os.Chdir(exeDir)
			godotenv.Load(filepath.Join(exeDir, ".env"))
		}
	}

	for _, a := range os.Args[1:] {
		if a == "--verbose" {
			k := syscall.NewLazyDLL("kernel32.dll")
			ret, _, _ := k.NewProc("AttachConsole").Call(uintptr(0xFFFFFFFF))
			if ret == 0 {
				k.NewProc("AllocConsole").Call()
			}
			stdout, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
			if err == nil {
				os.Stdout = stdout
			}

			stderr, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
			if err == nil {
				os.Stderr = stderr
			}
			break
		}
	}

	runNotifier := flag.Bool("run", false, "Ejecutar el worker de notificaciones")
	useWhatsmeow := flag.Bool("whatsmeow", false, "Usar whatsmeow en vez de WhatsApp API")
	validate := flag.Bool("validate", false, "Validar configuración y archivo Excel")
	installSchedule := flag.Bool("install-schedule", false, "Instalar tarea programada diaria (Task Scheduler)")
	removeSchedule := flag.Bool("remove-schedule", false, "Eliminar tarea programada")
	installDashboard := flag.Bool("install-dashboard", false, "Agregar dashboard al inicio de Windows")
	removeDashboard := flag.Bool("remove-dashboard", false, "Quitar dashboard del inicio de Windows")
	flag.Bool("verbose", false, "Mostrar logs en consola")

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

  --install-schedule        Instala tarea programada diaria a las 09:00.
                            Opcional: combinar con --whatsmeow.

  --remove-schedule         Elimina la tarea programada instalada.

  --install-dashboard       Agrega el dashboard al inicio de Windows
                            (se inicia automaticamente al iniciar sesion).

  --remove-dashboard        Quita el dashboard del inicio de Windows.

  --verbose                 Muestra los logs en consola (crea una si no hay).

 EJEMPLOS:
  notiflow                          Inicia el dashboard web
  notiflow --run                    Ejecuta notificaciones via WhatsApp API
  notiflow --run --whatsmeow        Ejecuta notificaciones via WhatsApp Personal
  notiflow --validate               Valida Excel y configuracion
  notiflow --install-schedule       Instala tarea diaria con API de Meta
  notiflow --install-schedule --whatsmeow   Instala tarea diaria con WhatsApp personal
  notiflow --remove-schedule        Desinstala la tarea programada
  notiflow --install-dashboard      Agrega dashboard al inicio de Windows
  notiflow --remove-dashboard       Quita dashboard del inicio de Windows`)
	}

	flag.Parse()

	logger.Init()

	if *installSchedule {
		scheduler.Install(*useWhatsmeow)
		return
	}
	if *removeSchedule {
		scheduler.Remove()
		return
	}
	if *installDashboard {
		scheduler.InstallDashboard()
		return
	}
	if *removeDashboard {
		scheduler.RemoveDashboard()
		return
	}

	serial, err := license.GetWindowSerialNumber()
	if err != nil {
		utils.ShowDialog("Error obteniendo serial", "Ocurrio un error obteniendo el serial, error: "+err.Error())
		logger.L.Sugar().Fatal(err.Error())
	}
	logger.L.Sugar().Infof("Numero de serial:%s", serial)

	if err := utils.WaitForInternet(context.Background()); err != nil {
		logger.L.Sugar().Infof("Error: %v", err.Error())
	} else {
		logger.L.Sugar().Info("Conexion a internet establecida")
	}

	active, err := license.ValidateLicense(serial)
	if err != nil {
		utils.ShowDialog("Error", err.Error())
		logger.L.Sugar().Fatal(err.Error())
	}
	if !active {
		utils.ShowDialog("Error", "Error verifique su conexion a internet, de lo contrario pongase en contacto")
		logger.L.Sugar().Fatal("Licencia no válida o expirada")
	}
	logger.L.Sugar().Infof("Licencia: %s", active)

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
