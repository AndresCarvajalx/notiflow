package notifier

import (
	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/model"
)

func FilterClients(clients []model.Client, days int) ([]model.Client, error) {
	var result []model.Client

	for _, client := range clients {
		if client.DaysOverdue < days {
			logger.L.Sugar().Debugf("Omitido (sin vencer): %s — %d días corridos", client.Name, client.DaysOverdue)
			continue
		}

		recentlyNotified, err := database.WasRecentlyNotified(client.Phone, client.Placa, days)
		if err != nil {
			return nil, err
		}

		if recentlyNotified {
			logger.L.Sugar().Infof("Omitido (notificado en los últimos %d días): %s", days, client.Name)
			continue
		}

		logger.L.Sugar().Infof("A notificar: %s — %d días corridos", client.Name, client.DaysOverdue)
		result = append(result, client)
	}

	return result, nil
}
