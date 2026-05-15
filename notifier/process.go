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
	filtered, err := FilterClients(clients, cfg.Scheduler.DiasVencimiento)
	if err != nil {
		return err
	}
	logger.L.Sugar().Infof("Clientes a notificar: %d", len(filtered))

	for _, client := range filtered {
		message := "Hola, buen día. Te escribimos para recordar que actualmente tienes una mora pendiente de *{{2}}* dias con nosotros referente a la placa *{{1}}* Agradecemos realizar el pago lo antes posible para evitar reportes negativos o posibles inconvenientes jurídicos derivados del incumplimiento. Si ya realizaste el pago, por favor envíanos el soporte para actualizar el estado de tu cuenta. Quedamos atentos. Gracias."

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
			client.Name,
			client.TipoTransaccion,
			client.Value,
			client.DaysOverdue,
			message,
			status,
			errorDetail,
		)
	}

	return nil
}
