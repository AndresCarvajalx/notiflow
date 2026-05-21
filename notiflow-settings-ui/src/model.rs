use serde::{Deserialize, Serialize};

#[derive(Default, Serialize, Deserialize)]
pub struct Data {
    pub excel: Excel,
    pub columnas: Columnas,
    pub whatsapp: WhatsApp,
    pub scheduler: Sheduler,
    pub server: Server,
}

#[derive(Default, Serialize, Deserialize)]
pub struct Excel {
    pub path: String,
    pub header_row: u8,
}

#[derive(Default, Serialize, Deserialize)]
pub struct Columnas {
    pub tipo_transaccion: String,
    pub cliente: String,
    pub placa: String,
    pub valor_actual: String,
    pub porcentaje_interes: String,
    pub valor_interes_mensual: String,
    pub vencimiento_interes: String,
    pub dias_corridos: String,
    pub valor_interes_vencido: String,
    pub saldo_actual: String,
    pub telefono: String,
}

#[derive(Default, Serialize, Deserialize)]
pub struct WhatsApp {
    pub token: String,
    pub phone_id: String,
    pub codigo_pais: String,
    pub mensaje: String,
}

#[derive(Default, Serialize, Deserialize)]
pub struct Sheduler {
    pub dias_vencimiento: u8,
}

#[derive(Default, Serialize, Deserialize)]
pub struct Server {
    pub port: i32,
}
