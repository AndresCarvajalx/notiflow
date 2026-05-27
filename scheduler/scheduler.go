package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Author>Notiflow</Author>
  </RegistrationInfo>
  <Triggers>
    <CalendarTrigger>
      <StartBoundary>2024-01-01T09:00:00</StartBoundary>
      <ScheduleByDay>
        <DaysInterval>1</DaysInterval>
      </ScheduleByDay>
    </CalendarTrigger>
  </Triggers>
  <Principals>
    <Principal id="Author">
      <LogonType>S4U</LogonType>
      <RunLevel>LeastPrivilege</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <StartWhenAvailable>true</StartWhenAvailable>
    <ExecutionTimeLimit>PT1H</ExecutionTimeLimit>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
  </Settings>
  <Actions>
    <Exec>
      <Command>%s</Command>
      <Arguments>%s</Arguments>
    </Exec>
  </Actions>
</Task>`, exePath, args)

	tmpFile := filepath.Join(os.TempDir(), "notiflow_task.xml")
	if err := os.WriteFile(tmpFile, []byte(xml), 0644); err != nil {
		logger.L.Sugar().Fatalf("Error creando XML de tarea: %v", err)
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/xml", tmpFile, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.L.Sugar().Fatalf("Error instalando tarea programada:\n%s\n%v", string(output), err)
	}

	logger.L.Sugar().Infof("Tarea programada '%s' instalada correctamente", taskName)
	fmt.Printf("\nTarea programada '%s' instalada.\n", taskName)
	fmt.Printf("   Ejecuta: %s %s\n", exePath, args)
	fmt.Printf("   Horario: Diario a las 09:00, ejecuta al encender si se paso de la hora\n")
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
	fmt.Printf("\nTarea programada '%s' eliminada.\n\n", taskName)
}

func InstallDashboard() {
	exePath, err := os.Executable()
	if err != nil {
		logger.L.Sugar().Fatalf("Error obteniendo ruta del ejecutable: %v", err)
	}

	reg := exec.Command("reg",
		"add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "Notiflow Dashboard",
		"/t", "REG_SZ",
		"/d", fmt.Sprintf(`"%s"`, exePath),
		"/f",
	)

	output, err := reg.CombinedOutput()
	if err != nil {
		logger.L.Sugar().Fatalf("Error instalando dashboard en inicio:\n%s\n%v", string(output), err)
	}

	logger.L.Sugar().Infof("Dashboard instalado en inicio de Windows")
	fmt.Printf("\nDashboard agregado al inicio de Windows.\n")
	fmt.Printf("   Ejecuta: %s\n", exePath)
	fmt.Printf("   Se iniciará automáticamente al iniciar sesión.\n")
	fmt.Printf("   Para quitarlo: notiflow --remove-dashboard\n\n")
}

func RemoveDashboard() {
	reg := exec.Command("reg",
		"delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "Notiflow Dashboard",
		"/f",
	)

	output, err := reg.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "does not exist") || strings.Contains(string(output), "no existe") {
			logger.L.Sugar().Warnf("El dashboard no está en el inicio de Windows")
			fmt.Printf("El dashboard no está instalado en el inicio.\n")
			return
		}
		logger.L.Sugar().Fatalf("Error quitando dashboard del inicio:\n%s\n%v", string(output), err)
	}

	logger.L.Sugar().Infof("Dashboard quitado del inicio de Windows")
	fmt.Printf("\nDashboard quitado del inicio de Windows.\n\n")
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
