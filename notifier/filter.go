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

		recentlyNotified, err := database.WasRecentlyNotified(client.Phone, client.Placa, days)
		if err != nil {
			return nil, nil, err
		}

		if recentlyNotified {
			logger.L.Sugar().Infof("Omitido (notificado en los últimos %d días): %s", days, client.Name)
			omissions = append(omissions, model.Omission{
				Client: client,
				Reason: fmt.Sprintf("Ya notificado en los últimos %d días", days),
			})
			continue
		}

		logger.L.Sugar().Infof("A notificar: %s — %d días corridos", client.Name, diasVencidos)
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
	now := time.Now().Truncate(24 * time.Hour)
	t = t.Truncate(24 * time.Hour)
	days := int(now.Sub(t).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
