package whatsapp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AndresCarvajalx/notiflow/logger"
	"github.com/AndresCarvajalx/notiflow/model"

	"github.com/go-resty/resty/v2"
)

func Send(token, phoneID, codigoPais string, cl model.Client) error {
	numero := strings.TrimLeft(cl.Phone, "0")

	if codigoPais != "" && !strings.HasPrefix(numero, codigoPais) {
		numero = codigoPais + numero
	}

	if cl.Placa == "" {
		cl.Placa = "N/A"
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                numero,
		"type":              "template",
		"template": map[string]interface{}{
			"name": "payment_reminder",
			"language": map[string]interface{}{
				"code": "es_CO",
			},
			"components": []map[string]interface{}{
				{
					"type": "body",
					"parameters": []map[string]interface{}{
						{
							"type": "text",
							"text": cl.Name,
						},
						{
							"type": "text",
							"text": strconv.Itoa(cl.DaysOverdue),
						},
						{
							"type": "text",
							"text": cl.TipoTransaccion,
						},
						{
							"type": "text",
							"text": cl.Placa,
						},
					},
				},
			},
		},
	}

	client := resty.New()
	resp, err := client.R().
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(fmt.Sprintf("https://graph.facebook.com/v25.0/%s/messages", phoneID))

	if err != nil {
		logger.L.Sugar().Errorf("Error occurend sending message to %s, %v", phoneID, err)
		return err
	}

	if resp.StatusCode() >= 300 {
		logger.L.Sugar().Errorf("whatsapp API error %d: %s", resp.StatusCode(), resp.String())
		return fmt.Errorf("whatsapp API error %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
