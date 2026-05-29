package config

import (
	"os"

	"github.com/AndresCarvajalx/notiflow/logger"
	"gopkg.in/yaml.v3"
)

var cfg *Config

type Config struct {
	Excel struct {
		Path      string `yaml:"path"`
		HeaderRow int    `yaml:"header_row"`
	} `yaml:"excel"`

	Columnas struct {
		TipoTransaccion     string `yaml:"tipo_transaccion"`
		Cliente             string `yaml:"cliente"`
		Placa               string `yaml:"placa"`
		ValorActual         string `yaml:"valor_actual"`
		PorcentajeInteres   string `yaml:"porcentaje_interes"`
		ValorInteresMensual string `yaml:"valor_interes_mensual"`
		VencimientoInteres  string `yaml:"vencimiento_interes"`
		DiasCorridos        string `yaml:"dias_corridos"`
		ValorInteresVencido string `yaml:"valor_interes_vencido"`
		SaldoActual         string `yaml:"saldo_actual"`
		Telefono            string `yaml:"telefono"`
	} `yaml:"columnas"`

	Whatsapp struct {
		Token      string `yaml:"token"`
		PhoneID    string `yaml:"phone_id"`
		CodigoPais string `yaml:"codigo_pais"`
		Mensaje    string `yaml:"mensaje"`
	} `yaml:"whatsapp"`

	Scheduler struct {
		DiasVencimiento int `yaml:"dias_vencimiento"`
	} `yaml:"scheduler"`

	Throttle struct {
		DelayMinSegundos int `yaml:"delay_min_segundos"`
		DelayMaxSegundos int `yaml:"delay_max_segundos"`
	} `yaml:"throttle"`

	Server struct {
		Port string `yaml:"port" `
	} `yaml:"server" `
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(file, c)
	if err != nil {
		return nil, err
	}

	cfg = c
	return c, nil
}

func Get() *Config {
	if cfg == nil {
		logger.L.Panic("Config not loaded")
		panic("Config not loaded")
	}
	return cfg
}
