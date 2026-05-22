package wmeow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"

	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/database"
	"github.com/AndresCarvajalx/notiflow/excel"
	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/model"
	"github.com/AndresCarvajalx/notiflow/notifier"
	"github.com/AndresCarvajalx/notiflow/utils"
)

var client *whatsmeow.Client

func connect() error {
	dbLog := waLog.Stdout("Database", "ERROR", true)
	ctx := context.Background()

	container, err := sqlstore.New(ctx, "sqlite", "file:whatsmeow.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbLog)
	if err != nil {
		return fmt.Errorf("error abriendo store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo device: %w", err)
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)

	connected := make(chan struct{}, 1)
	loggedOut := make(chan struct{}, 1)

	client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.Connected:
			logger.L.Sugar().Info("WhatsApp conectado y listo")
			select {
			case connected <- struct{}{}:
			default:
			}
		case *events.LoggedOut:
			logger.L.Sugar().Warn("Sesión cerrada, se necesita vincular de nuevo")
			select {
			case loggedOut <- struct{}{}:
			default:
			}
		}
	})

	if client.Store.ID == nil {
		qrCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		qrChan, _ := client.GetQRChannel(qrCtx)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("error conectando: %w", err)
		}

		qrCodeCh := make(chan string, 1)
		scanned := make(chan struct{})

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					select {
					case qrCodeCh <- evt.Code:
					default:
					}
				} else if evt.Event == "success" {
					close(scanned)
					return
				}
			}
		}()

		var code string
		select {
		case code = <-qrCodeCh:
		case <-time.After(2 * time.Minute):
			client.Disconnect()
			return fmt.Errorf("no se genero el codigo QR en el tiempo limite")
		}

		qr, err := qrcode.New(code, qrcode.Medium)
		if err != nil {
			client.Disconnect()
			return fmt.Errorf("error generando QR: %w", err)
		}
		png, err := qr.PNG(256)
		if err != nil {
			client.Disconnect()
			return fmt.Errorf("error generando imagen QR: %w", err)
		}

		if err := utils.ShowQRDialog(png, "Notiflow - Escanea el QR"); err != nil {
			logger.L.Sugar().Warnf("Error mostrando dialogo QR: %v", err)
		}

		select {
		case <-scanned:
			logger.L.Sugar().Info("QR escaneado correctamente")
		case <-time.After(10 * time.Second):
			client.Disconnect()
			return fmt.Errorf("no se detecto el escaneo del QR")
		}
	} else {
		if err := client.Connect(); err != nil {
			return fmt.Errorf("error conectando: %w", err)
		}
	}

	select {
	case <-connected:
		logger.L.Sugar().Info("Listo para enviar mensajes")
	case <-loggedOut:
		client.Disconnect()
		return fmt.Errorf("sesión de WhatsApp cerrada o expirada, ejecuta notiflow en una terminal para escanear el QR de nuevo")
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout esperando conexión con WhatsApp")
	}

	return nil
}

func disconnect() {
	if client != nil {
		client.Disconnect()
	}
}

func sendMessage(cl model.Client, message string) error {
	cfg := config.Get()

	numero := strings.TrimLeft(cl.Phone, "0")
	if cfg.Whatsapp.CodigoPais != "" && !strings.HasPrefix(numero, cfg.Whatsapp.CodigoPais) {
		numero = cfg.Whatsapp.CodigoPais + numero
	}

	jid, err := types.ParseJID(numero + "@s.whatsapp.net")
	if err != nil {
		return fmt.Errorf("JID inválido para %s: %w", cl.Phone, err)
	}

	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}

	_, err = client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		logger.L.Sugar().Warnf("Reintentando envío a %s...", cl.Phone)
		time.Sleep(2 * time.Second)
		_, err = client.SendMessage(context.Background(), jid, msg)
	}

	return err
}

func Run() error {
	cfg := config.Get()

	if err := connect(); err != nil {
		return err
	}
	defer disconnect()

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
	filtered, omissions, err := notifier.FilterClients(clients, cfg.Scheduler.DiasVencimiento)
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

	for _, cl := range filtered {
		if cl.Placa == "" {
			cl.Placa = "N/A"
		}

		message := buildMessage(cl)

		err := sendMessage(cl, message)

		status := "enviado"
		var errorDetail *string

		if err != nil {
			status = "error"
			msg := err.Error()
			errorDetail = &msg
			logger.L.Sugar().Errorf("Error enviando a %s: %v", cl.Name, err)
		} else {
			logger.L.Sugar().Infof("Enviado a %s (%s)", cl.Name, cl.Phone)
		}

		_ = database.CreateNotification(
			cl.Phone,
			cl.Placa,
			cl.Name,
			cl.TipoTransaccion,
			cl.Value,
			cl.DaysOverdue,
			message,
			status,
			errorDetail,
		)

		time.Sleep(2 * time.Second)
	}

	return nil
}

func buildMessage(cl model.Client) string {
	cfg := config.Get()
	msg := cfg.Whatsapp.Mensaje

	nombre := cl.Name
	if nombre == "" {
		nombre = "Cliente"
	}

	placa := cl.Placa
	if placa == "" {
		placa = "N/A"
	}

	tipo := cl.TipoTransaccion
	if tipo == "" {
		tipo = "N/A"
	}

	dias := fmt.Sprintf("%d", cl.DaysOverdue)
	if cl.DaysOverdue == 0 {
		dias = "N/A"
	}

	valor := fmt.Sprintf("$ %.2f", cl.Value)
	if cl.Value == 0 {
		valor = "N/A"
	}

	saldo := fmt.Sprintf("$ %.2f", cl.SaldoActual)
	if cl.SaldoActual == 0 {
		saldo = "N/A"
	}

	msg = strings.ReplaceAll(msg, "{cliente}", nombre)
	msg = strings.ReplaceAll(msg, "{dias}", dias)
	msg = strings.ReplaceAll(msg, "{placa}", placa)
	msg = strings.ReplaceAll(msg, "{valor}", valor)
	msg = strings.ReplaceAll(msg, "{saldo}", saldo)
	msg = strings.ReplaceAll(msg, "{tipo}", tipo)

	return msg
}
