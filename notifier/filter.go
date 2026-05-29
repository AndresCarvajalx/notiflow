package notifier

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/model"
)

func FilterClients(clients []model.Client, days int) ([]model.Client, []model.Omission, error) {
	var toNotify []model.Client
	var omissions []model.Omission

	for _, client := range clients {
		if d := diasDesdeVencimiento(client.VencimientoInteres); d >= 0 {
			client.DaysOverdue = d
		}
		diasVencidos := client.DaysOverdue

		if diasVencidos < days {
			logger.L.Sugar().Debugf("Omitido (sin vencer): %s — %d días corridos", client.Name, diasVencidos)
			omissions = append(omissions, model.Omission{
				Client: client,
				Reason: fmt.Sprintf("Días vencidos (%d) menor al umbral (%d)", diasVencidos, days),
			})
			continue
		}

		cicloActual := diasVencidos / days

		if cicloActual < 1 {
			logger.L.Sugar().Debugf("Omitido (sin vencer): %s — %d días corridos", client.Name, diasVencidos)
			omissions = append(omissions, model.Omission{
				Client: client,
				Reason: fmt.Sprintf("Días vencidos (%d) menor al umbral (%d)", diasVencidos, days),
			})
			continue
		}

		ultimoCiclo, err := database.GetUltimoCiclo(client.Phone, client.Placa)
		if err != nil {
			return nil, nil, err
		}

		if cicloActual <= ultimoCiclo {
			logger.L.Sugar().Infof("Omitido (ciclo %d ya notificado): %s — %d días", ultimoCiclo, client.Name, diasVencidos)
			omissions = append(omissions, model.Omission{
				Client: client,
				Reason: fmt.Sprintf("Ya notificado en ciclo %d (días: %d)", ultimoCiclo, diasVencidos),
			})
			continue
		}

		logger.L.Sugar().Infof("A notificar: %s — %d días (ciclo %d)", client.Name, diasVencidos, cicloActual)
		toNotify = append(toNotify, client)
	}

	return toNotify, omissions, nil
}

func diasDesdeVencimiento(dateStr string) int {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return -1
	}

	var t time.Time
	var err error

	formatos := []string{
		"02/01/2006",
		"02-01-2006",
		"2006-01-02",
		"02/01/2006 15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, f := range formatos {
		t, err = time.Parse(f, dateStr)
		if err == nil {
			goto calc
		}
	}

	if n, err := strconv.Atoi(dateStr); err == nil && n > 40000 {
		epoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		t = epoch.AddDate(0, 0, n)
		goto calc
	}

	return -1

calc:
	now := time.Now()
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.Date()
	hoy := time.Date(y1, m1, d1, 0, 0, 0, 0, now.Location())
	venc := time.Date(y2, m2, d2, 0, 0, 0, 0, now.Location())
	days := int(hoy.Sub(venc).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
