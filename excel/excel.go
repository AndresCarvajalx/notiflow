package excel

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/model"
	"github.com/xuri/excelize/v2"
)

func ReadClients(path string) ([]model.Client, error) {
	tempFile := filepath.Join(os.TempDir(), "notiflow_temp.xlsx")

	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(tempFile)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	f, err := excelize.OpenFile(tempFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}

	cfg := config.Get()
	headerRow := cfg.Excel.HeaderRow
	if headerRow == 0 {
		headerRow = 1
	}
	headerIdx := headerRow - 1

	if len(rows) <= headerIdx {
		return nil, fmt.Errorf("el archivo no tiene suficientes filas para el header en fila %d", headerRow)
	}

	colIdx := buildColumnIndex(rows[headerIdx])

	cols := cfg.Columnas

	idxTipo := resolve(colIdx, cols.TipoTransaccion)
	idxCliente := resolve(colIdx, cols.Cliente)
	idxPlaca := resolve(colIdx, cols.Placa)
	idxValor := resolve(colIdx, cols.ValorActual)
	idxPct := resolve(colIdx, cols.PorcentajeInteres)
	idxIntMens := resolve(colIdx, cols.ValorInteresMensual)
	idxVenc := resolve(colIdx, cols.VencimientoInteres)
	idxDias := resolve(colIdx, cols.DiasCorridos)
	idxIntVenc := resolve(colIdx, cols.ValorInteresVencido)
	idxSaldo := resolve(colIdx, cols.SaldoActual)
	idxTel := resolve(colIdx, cols.Telefono)

	var clients []model.Client

	for i, row := range rows {
		if i <= headerIdx {
			continue
		}

		name := getCell(row, idxCliente)
		phone := CleanPhone(getCell(row, idxTel))

		if name == "" || phone == "" {
			continue
		}

		client := model.Client{
			TipoTransaccion:     getCell(row, idxTipo),
			Name:                name,
			Placa:               getCell(row, idxPlaca),
			Value:               parseFloat(getCell(row, idxValor)),
			PorcentajeInteres:   parseFloat(getCell(row, idxPct)),
			ValorInteresMensual: parseFloat(getCell(row, idxIntMens)),
			VencimientoInteres:  getCell(row, idxVenc),
			DaysOverdue:         parseInt(getCell(row, idxDias)),
			ValorInteresVencido: parseFloat(getCell(row, idxIntVenc)),
			SaldoActual:         parseFloat(getCell(row, idxSaldo)),
			Phone:               phone,
		}

		clients = append(clients, client)
	}

	return clients, nil
}

func buildColumnIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[normalize(h)] = i
	}
	return idx
}

func resolve(colIdx map[string]int, name string) int {
	if i, ok := colIdx[normalize(name)]; ok {
		return i
	}
	return -1
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func getCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func parseInt(s string) int {
	if dot := strings.Index(s, "."); dot != -1 {
		s = s[:dot]
	}
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func CleanPhone(phone string) string {
	var result strings.Builder
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			result.WriteRune(c)
		}
	}
	return result.String()
}
