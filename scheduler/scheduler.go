package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AndresCarvajalx/notiflow/logger"
)

const taskName = "Notiflow"

func Install(useWhatsmeow bool) {
	exePath, err := os.Executable()
	if err != nil {
		logger.L.Sugar().Fatalf("Error obteniendo ruta del ejecutable: %v", err)
	}

	args := "--run"
	if useWhatsmeow {
		args += " --whatsmeow"
	}

	cmd := fmt.Sprintf(`"%s" %s`, exePath, args)
	schtasks := exec.Command("schtasks",
		"/create", "/tn", taskName,
		"/tr", cmd,
		"/sc", "daily",
		"/st", "09:00",
		"/f",
		"/it",
	)

	output, err := schtasks.CombinedOutput()
	if err != nil {
		logger.L.Sugar().Fatalf("Error instalando tarea programada:\n%s\n%v", string(output), err)
	}

	logger.L.Sugar().Infof("Tarea programada '%s' instalada correctamente (diaria a las 09:00)", taskName)
	fmt.Printf("\nTarea programada '%s' instalada.\n", taskName)
	fmt.Printf("   Ejecuta: %s %s\n", exePath, args)
	fmt.Printf("   Horario: Diario a las 09:00\n")
	fmt.Printf("   Para desinstalar: notiflow --remove-schedule\n\n")
}

func Remove() {
	schtasks := exec.Command("schtasks",
		"/delete", "/tn", taskName,
		"/f",
	)

	output, err := schtasks.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "does not exist") || strings.Contains(string(output), "no existe") {
			logger.L.Sugar().Warnf("La tarea '%s' no existe", taskName)
			fmt.Printf("La tarea programada '%s' no existe.\n", taskName)
			return
		}
		logger.L.Sugar().Fatalf("Error eliminando tarea programada:\n%s\n%v", string(output), err)
	}

	logger.L.Sugar().Infof("Tarea programada '%s' eliminada", taskName)
	fmt.Printf("\n✅ Tarea programada '%s' eliminada.\n\n", taskName)
}

func Status() bool {
	schtasks := exec.Command("schtasks",
		"/query", "/tn", taskName,
		"/fo", "csv", "/nh",
	)

	output, err := schtasks.CombinedOutput()
	if err != nil {
		return false
	}

	return len(output) > 0 && !strings.Contains(string(output), "no existe")
}
