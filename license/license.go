package license

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ValidateLicense(serialNumber string) (bool, error) {
	token := os.Getenv("LICENSE_TOKEN")
	const URL = "https://api.github.com/repos/AndresCarvajalx/notiflow-license/contents/data.json"
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return false, err

	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.raw+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("error consultando licencia: %d", resp.StatusCode)
	}

	var data struct {
		Licenses []struct {
			Client     string `json:"client"`
			Active     bool   `json:"active"`
			Expiration string `json:"expiration"`
		} `json:"licenses"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, err
	}

	for _, l := range data.Licenses {
		if l.Client == serialNumber && l.Active {
			expiration, err := time.Parse("2006-01-02", l.Expiration)
			if err != nil {
				return false, err
			}
			expiration = expiration.Add(96 * time.Hour)
			return time.Now().Before(expiration), nil
		}
	}

	return false, nil
}

func GetWindowSerialNumber() (string, error) {
	cmd := exec.Command("wmic", "bios", "get", "serialnumber")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line != "" && line != "SerialNumber" {
			return line, nil
		}
	}

	return "", fmt.Errorf("serial not found")
}
