package notifier

import (
	"github.com/AndresCarvajalx/notiflow/logger"

	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/excel"
	"github.com/AndresCarvajalx/notiflow/whatsapp"
)

func Run() error {
	cfg := config.Get()

	logger.L.Sugar().Infof("Leyendo Excel: %s", cfg.Excel.Path)
	clients, err := excel.ReadClients(cfg.Excel.Path)
	if err != nil {
		return err
	}
	logger.L.Sugar().Infof("Total clientes leidos: %d", len(clients))
	for _, c := range clients {
		logger.L.Sugar().Infof("  %-30s | Tel: %s | Dias: %d", c.Name, c.Phone, c.DaysOverdue)
	}

	logger.L.Sugar().Infof("Filtrando clientes con mas de %d dias vencidos...", cfg.Scheduler.DiasVencimiento)
	filtered, omissions, err := FilterClients(clients, cfg.Scheduler.DiasVencimiento)
	if err != nil {
		return err
	}
	logger.L.Sugar().Infof("Clientes a notificar: %d, omitidos: %d", len(filtered), len(omissions))

	for _, om := range omissions {
		_ = database.CreateNotification(
			om.Client.Phone,
			om.Client.Placa,
			om.Client.Name,
			om.Client.TipoTransaccion,
			om.Client.Value,
			om.Client.DaysOverdue,
			"",
			"omitido",
			&om.Reason,
		)
	}

	for _, client := range filtered {
		err := whatsapp.Send(
			cfg.Whatsapp.Token,
			cfg.Whatsapp.PhoneID,
			cfg.Whatsapp.CodigoPais,
			client,
		)

		status := "enviado"
		var errorDetail *string

		if err != nil {
			status = "error"
			msg := err.Error()
			errorDetail = &msg
			logger.L.Sugar().Errorf("Error enviando a %s: %v", client.Name, err)
		} else {
			logger.L.Sugar().Infof("Enviado a %s (%s)", client.Name, client.Phone)
		}

		_ = database.CreateNotification(
			client.Phone,
			client.Placa,
			client.Name,
			client.TipoTransaccion,
			client.Value,
			client.DaysOverdue,
			"",
			status,
			errorDetail,
		)
	}

	return nil
}
