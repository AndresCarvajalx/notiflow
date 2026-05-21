use std::collections::HashMap;
use std::fs;

use calamine::{Data, Reader, Xlsx, open_workbook};

use crate::model;

pub struct ValidationResult {
    pub is_valid: bool,
    pub file_exists: bool,
    pub columns_missing: Vec<String>,
    pub total_rows: u32,
    pub rows_with_phone: u32,
    pub rows_without_phone: u32,
    pub sample_phone: String,
    pub errors: Vec<String>,
    pub warnings: Vec<String>,
}

fn cell_string(cell: &Data) -> String {
    match cell {
        Data::String(s) => s.clone(),
        Data::Empty => String::new(),
        other => other.to_string(),
    }
}

pub fn validate(data: &model::Data) -> ValidationResult {
    let mut r = ValidationResult {
        is_valid: true,
        file_exists: false,
        columns_missing: Vec::new(),
        total_rows: 0,
        rows_with_phone: 0,
        rows_without_phone: 0,
        sample_phone: String::new(),
        errors: Vec::new(),
        warnings: Vec::new(),
    };

    let path = &data.excel.path;
    if path.is_empty() {
        r.errors.push("La ruta del archivo Excel está vacía".into());
        r.is_valid = false;
        return r;
    }

    match fs::metadata(path) {
        Ok(m) if m.is_file() => r.file_exists = true,
        _ => {
            r.errors.push(format!("Archivo no encontrado: {}", path));
            r.is_valid = false;
            return r;
        }
    }

    let mut workbook: Xlsx<_> = match open_workbook(path) {
        Ok(wb) => wb,
        Err(e) => {
            r.errors.push(format!("Error abriendo Excel: {}", e));
            r.is_valid = false;
            return r;
        }
    };

    let sheet_names = workbook.sheet_names();

    if sheet_names.is_empty() {
        r.errors.push("El archivo Excel no tiene hojas".into());
        r.is_valid = false;
        return r;
    }

    let range = match workbook.worksheet_range(&sheet_names[0]) {
        Ok(rng) => rng,
        Err(e) => {
            r.errors
                .push(format!("Error leyendo hoja '{}': {}", sheet_names[0], e));
            r.is_valid = false;
            return r;
        }
    };

    let header_row = if data.excel.header_row == 0 {
        1
    } else {
        data.excel.header_row as usize
    };

    let rows: Vec<Vec<Data>> = range.rows().map(|r| r.to_vec()).collect();

    if rows.len() < header_row {
        r.errors.push(format!(
            "El archivo tiene {} filas, se necesita al menos {}",
            rows.len(),
            header_row
        ));
        r.is_valid = false;
        return r;
    }

    let mut actual_headers: HashMap<String, usize> = HashMap::new();
    for (i, cell) in rows[header_row - 1].iter().enumerate() {
        let h = cell_string(cell).trim().to_lowercase();
        if !h.is_empty() {
            actual_headers.insert(h, i);
        }
    }

    let cols: Vec<(&str, &str)> = vec![
        ("Tipo Transacción", &data.columnas.tipo_transaccion),
        ("Cliente", &data.columnas.cliente),
        ("Placa", &data.columnas.placa),
        ("Valor Actual", &data.columnas.valor_actual),
        ("% Interés", &data.columnas.porcentaje_interes),
        (
            "Valor Interés Mensual",
            &data.columnas.valor_interes_mensual,
        ),
        ("Vencimiento Interés", &data.columnas.vencimiento_interes),
        ("Días Corridos", &data.columnas.dias_corridos),
        (
            "Valor Interés Vencido",
            &data.columnas.valor_interes_vencido,
        ),
        ("Saldo Actual", &data.columnas.saldo_actual),
        ("Teléfono", &data.columnas.telefono),
    ];

    for (label, col_name) in &cols {
        let normal = col_name.trim().to_lowercase();
        if !actual_headers.contains_key(&normal) {
            r.columns_missing
                .push(format!("{} ('{}')", label, col_name));
        }
    }

    if !r.columns_missing.is_empty() {
        r.warnings.push(format!(
            "Columnas no encontradas: {}",
            r.columns_missing.join(", ")
        ));
    }

    let phone_col = cols
        .iter()
        .position(|(label, _)| *label == "Teléfono")
        .and_then(|_| {
            let col_name = &data.columnas.telefono;
            let normal = col_name.trim().to_lowercase();
            actual_headers.get(&normal).copied()
        });

    for i in header_row..rows.len() {
        let row = &rows[i];
        if row.is_empty() || row.iter().all(|c| matches!(c, Data::Empty)) {
            continue;
        }
        r.total_rows += 1;

        if let Some(col) = phone_col {
            let phone = clean_phone(&cell_string(row.get(col).unwrap_or(&Data::Empty)));
            if phone.is_empty() {
                r.rows_without_phone += 1;
            } else {
                r.rows_with_phone += 1;
                if r.sample_phone.is_empty() {
                    r.sample_phone = phone;
                }
            }
        }
    }

    if r.total_rows == 0 {
        r.warnings
            .push("No hay filas de datos después del encabezado".into());
    }

    if phone_col.is_some() && r.rows_with_phone == 0 {
        r.warnings
            .push("Ninguna fila tiene un número de teléfono válido".into());
    }

    if r.rows_without_phone > 0 {
        r.warnings.push(format!(
            "{} fila(s) sin teléfono (serán omitidas)",
            r.rows_without_phone
        ));
    }

    if phone_col.is_none() {
        r.errors.push("Columna de teléfono no encontrada".into());
        r.is_valid = false;
    }

    r
}

fn clean_phone(s: &str) -> String {
    s.chars().filter(|c| c.is_ascii_digit()).collect()
}

impl ValidationResult {
    pub fn summary(&self) -> String {
        let mut s = String::new();

        s.push_str("   VALIDACIÓN DE CONFIGURACIÓN\n");

        if !self.errors.is_empty() {
            s.push_str("❌ ERRORES:\n");
            for e in &self.errors {
                s.push_str(&format!("  • {}\n", e));
            }
            s.push('\n');
        }

        if !self.warnings.is_empty() {
            s.push_str("ADVERTENCIAS:\n");
            for w in &self.warnings {
                s.push_str(&format!("  • {}\n", w));
            }
            s.push('\n');
        }

        s.push_str("📊 RESUMEN:\n");
        s.push_str(&format!(
            "  Archivo:        {}\n",
            if self.file_exists {
                "✅ Encontrado"
            } else {
                "❌ No encontrado"
            }
        ));
        s.push_str(&format!(
            "  Columnas OK:    {}/11\n",
            11 - self.columns_missing.len()
        ));
        s.push_str(&format!("  Filas de datos: {}\n", self.total_rows));
        s.push_str(&format!("  Con teléfono:   {}\n", self.rows_with_phone));
        s.push_str(&format!("  Sin teléfono:   {}\n", self.rows_without_phone));
        if !self.sample_phone.is_empty() {
            s.push_str(&format!("  Tel. ejemplo:   {}\n", self.sample_phone));
        }
        s.push('\n');

        if self.is_valid {
            s.push_str("✅ TODO CORRECTO\n");
        } else {
            s.push_str("❌ HAY ERRORES — Corrige antes de ejecutar\n");
        }

        s
    }
}
