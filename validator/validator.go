package validator

import (
	"fmt"
	"os"
	"strings"

	"github.com/AndresCarvajalx/notiflow/config"
	"github.com/AndresCarvajalx/notiflow/excel"
	"github.com/xuri/excelize/v2"
)

type Result struct {
	Valid            bool
	FileExists       bool
	ColumnsMissing   []string
	TotalRows        int
	RowsWithPhone    int
	RowsWithoutPhone int
	SamplePhone      string
	Errors           []string
	Warnings         []string
}

func Run(cfg *config.Config) *Result {
	r := &Result{Valid: true}

	info, err := os.Stat(cfg.Excel.Path)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("Archivo no encontrado: %s", cfg.Excel.Path))
		r.Valid = false
		return r
	}
	if info.IsDir() {
		r.Errors = append(r.Errors, fmt.Sprintf("La ruta es un directorio: %s", cfg.Excel.Path))
		r.Valid = false
		return r
	}
	r.FileExists = true

	f, err := excelize.OpenFile(cfg.Excel.Path)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("Error abriendo Excel: %v", err))
		r.Valid = false
		return r
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("Error leyendo hoja '%s': %v", sheet, err))
		r.Valid = false
		return r
	}

	headerRow := cfg.Excel.HeaderRow
	if headerRow == 0 {
		headerRow = 1
	}
	headerIdx := headerRow - 1

	if len(rows) <= headerIdx {
		r.Errors = append(r.Errors, fmt.Sprintf("El archivo no tiene suficientes filas (tiene %d, necesita al menos %d)", len(rows), headerRow))
		r.Valid = false
		return r
	}

	actualHeaders := make(map[string]int)
	for i, h := range rows[headerIdx] {
		actualHeaders[strings.ToLower(strings.TrimSpace(h))] = i
	}

	cols := cfg.Columnas
	columnFields := map[string]string{
		"Tipo Transacción":      cols.TipoTransaccion,
		"Cliente":               cols.Cliente,
		"Placa":                 cols.Placa,
		"Valor Actual":          cols.ValorActual,
		"% Interés":             cols.PorcentajeInteres,
		"Valor Interés Mensual": cols.ValorInteresMensual,
		"Vencimiento Interés":   cols.VencimientoInteres,
		"Días Corridos":         cols.DiasCorridos,
		"Valor Interés Vencido": cols.ValorInteresVencido,
		"Saldo Actual":          cols.SaldoActual,
		"Teléfono":              cols.Telefono,
	}

	for label, colName := range columnFields {
		normal := strings.ToLower(strings.TrimSpace(colName))
		if _, ok := actualHeaders[normal]; !ok {
			r.ColumnsMissing = append(r.ColumnsMissing, fmt.Sprintf("%s ('%s')", label, colName))
		}
	}

	if len(r.ColumnsMissing) > 0 {
		r.Warnings = append(r.Warnings, fmt.Sprintf("Columnas no encontradas en el Excel: %s", strings.Join(r.ColumnsMissing, ", ")))
	}

	phoneCol, phoneFound := actualHeaders[strings.ToLower(strings.TrimSpace(cols.Telefono))]
	nameCol, _ := actualHeaders[strings.ToLower(strings.TrimSpace(cols.Cliente))]

	for i := headerIdx + 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}
		r.TotalRows++

		if phoneFound {
			phone := excel.CleanPhone(excel.GetCell(row, phoneCol))
			if phone == "" {
				r.RowsWithoutPhone++
			} else {
				r.RowsWithPhone++
				if r.SamplePhone == "" {
					r.SamplePhone = phone
				}
			}
		}

		if nameCol >= 0 && nameCol < len(row) && strings.TrimSpace(row[nameCol]) == "" && phoneFound {
			phone := excel.CleanPhone(excel.GetCell(row, phoneCol))
			if phone != "" {
				r.Warnings = append(r.Warnings, fmt.Sprintf("Fila %d: tiene teléfono pero el nombre está vacío", i+1))
			}
		}
	}

	if r.TotalRows == 0 {
		r.Warnings = append(r.Warnings, "No hay filas de datos después de la fila de encabezado")
	}

	if phoneFound && r.RowsWithPhone == 0 {
		r.Warnings = append(r.Warnings, "Ninguna fila tiene un número de teléfono válido")
	}

	if r.RowsWithoutPhone > 0 {
		r.Warnings = append(r.Warnings, fmt.Sprintf("%d filas sin teléfono válido (serán omitidas al notificar)", r.RowsWithoutPhone))
	}

	if !phoneFound {
		r.Errors = append(r.Errors, "Columna de teléfono no encontrada — no se podrá notificar a nadie")
		r.Valid = false
	}

	return r
}

func (r *Result) String() string {
	var b strings.Builder
	b.WriteString("VALIDACIÓN DE CONFIGURACIÓN\n")

	if len(r.Errors) > 0 {
		b.WriteString("ERRORES:\n")
		for _, e := range r.Errors {
			b.WriteString(fmt.Sprintf("  • %s\n", e))
		}
		b.WriteString("\n")
	}

	if len(r.Warnings) > 0 {
		b.WriteString("ADVERTENCIAS:\n")
		for _, w := range r.Warnings {
			b.WriteString(fmt.Sprintf("  • %s\n", w))
		}
		b.WriteString("\n")
	}

	b.WriteString("RESUMEN:\n")
	b.WriteString(fmt.Sprintf("  Archivo:        %s\n", map[bool]string{true: "Encontrado", false: "No encontrado"}[r.FileExists]))
	b.WriteString(fmt.Sprintf("  Columnas OK:    %d/11\n", 11-len(r.ColumnsMissing)))
	b.WriteString(fmt.Sprintf("  Filas de datos: %d\n", r.TotalRows))
	b.WriteString(fmt.Sprintf("  Con teléfono:   %d\n", r.RowsWithPhone))
	b.WriteString(fmt.Sprintf("  Sin teléfono:   %d\n", r.RowsWithoutPhone))
	if r.SamplePhone != "" {
		b.WriteString(fmt.Sprintf("  Tel. ejemplo:   %s\n", r.SamplePhone))
	}
	b.WriteString("\n")

	if r.Valid {
		b.WriteString("TODO CORRECTO — La configuración es válida\n")
	} else {
		b.WriteString("HAY ERRORES — Corrige los problemas antes de ejecutar\n")
	}

	return b.String()
}
