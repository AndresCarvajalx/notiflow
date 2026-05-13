package notifier

import (
	"fmt"
	"strings"

	"github.com/AndresCarvajalx/notiflow/logger"

	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/excel"
	"github.com/AndresCarvajalx/notiflow/model"
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
		message := buildMessage(cfg.Mensaje, client)

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

func buildMessage(plantilla string, c model.Client) string {
	if plantilla == "" {
		plantilla = "Hola {nombre}, su compromiso ({descripcion}) tiene {dias_vencidos} días vencidos. Saldo: ${saldo:,.0f}. Comuníquese con nosotros."
	}

	r := strings.NewReplacer(
		"{nombre}", c.Name,
		"{descripcion}", c.TipoTransaccion,
		"{dias_vencidos}", fmt.Sprintf("%d", c.DaysOverdue),
		"{valor}", fmt.Sprintf("%.0f", c.Value),
		"{saldo}", fmt.Sprintf("%.0f", c.SaldoActual),
		"{vencimiento}", c.VencimientoInteres,
		"{placa}", c.Placa,
		"{interes}", fmt.Sprintf("%.0f", c.ValorInteresMensual),
	)

	return r.Replace(plantilla)
}
