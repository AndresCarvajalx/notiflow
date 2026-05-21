package notifier

import (
	"fmt"

	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/model"
)

func FilterClients(clients []model.Client, days int) ([]model.Client, []model.Omission, error) {
	var toNotify []model.Client
	var omissions []model.Omission

	for _, client := range clients {
		if client.DaysOverdue < days {
			logger.L.Sugar().Debugf("Omitido (sin vencer): %s — %d días corridos", client.Name, client.DaysOverdue)
			omissions = append(omissions, model.Omission{
				Client: client,
				Reason: fmt.Sprintf("Días vencidos (%d) menor al umbral (%d)", client.DaysOverdue, days),
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

		logger.L.Sugar().Infof("A notificar: %s — %d días corridos", client.Name, client.DaysOverdue)
		toNotify = append(toNotify, client)
	}

	return toNotify, omissions, nil
}
