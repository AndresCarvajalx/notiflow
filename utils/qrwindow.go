package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ShowQRDialog(pngData []byte, title string) error {
	tmpDir := os.TempDir()
	path := filepath.Join(tmpDir, "notiflow_qr.png")

	if err := os.WriteFile(path, pngData, 0644); err != nil {
		return fmt.Errorf("error escribiendo QR: %w", err)
	}

	pngPath := strings.ReplaceAll(path, "\\", "\\\\")

	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$f = New-Object System.Windows.Forms.Form
$f.Text = "%s"
$f.Size = New-Object System.Drawing.Size(360, 430)
$f.StartPosition = "CenterScreen"
$f.TopMost = $true
$f.FormBorderStyle = "FixedDialog"
$f.MaximizeBox = $false
$f.MinimizeBox = $false
$pb = New-Object System.Windows.Forms.PictureBox
$pb.SizeMode = "Zoom"
$pb.Image = [System.Drawing.Image]::FromFile("%s")
$pb.Size = New-Object System.Drawing.Size(300, 300)
$pb.Location = New-Object System.Drawing.Point(30, 20)
$l = New-Object System.Windows.Forms.Label
$l.Text = "Escanea el codigo QR con WhatsApp"
$l.TextAlign = "MiddleCenter"
$l.Font = New-Object System.Drawing.Font("Segoe UI", 10)
$l.Size = New-Object System.Drawing.Size(300, 30)
$l.Location = New-Object System.Drawing.Point(30, 330)
$f.Controls.Add($pb)
$f.Controls.Add($l)
[void]$f.ShowDialog()
`, title, pngPath)

	cmd := exec.Command("powershell",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-WindowStyle", "Hidden",
		"-Command", script,
	)
	return cmd.Run()
}
