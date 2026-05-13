package model

type Client struct {
	TipoTransaccion     string
	Name                string
	Placa               string
	Value               float64
	PorcentajeInteres   float64
	ValorInteresMensual float64
	VencimientoInteres  string
	DaysOverdue         int
	ValorInteresVencido float64
	SaldoActual         float64
	Phone               string
}
