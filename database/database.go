package database

import (
	"database/sql"
	"time"

	"github.com/AndresCarvajalx/notiflow/logger"
	_ "modernc.org/sqlite"
)

var schema = `
CREATE TABLE IF NOT EXISTS configuracion (
    clave        TEXT PRIMARY KEY,
    valor        TEXT,
    descripcion  TEXT,
    updated_at   TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS columnas_excel (
    campo_interno         TEXT PRIMARY KEY,
    nombre_columna_excel  TEXT,
    descripcion           TEXT
);

CREATE TABLE IF NOT EXISTS notificaciones (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    telefono         TEXT NOT NULL,
	placa            TEXT,
    nombre           TEXT,
    descripcion      TEXT,
    valor            REAL,
    dias_vencidos    INTEGER,
    mensaje_enviado  TEXT,
    estado           TEXT DEFAULT 'pendiente'
                      CHECK (estado IN ('pendiente','enviado','error','omitido')),
    error_detalle    TEXT,
    fecha_envio      TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_notificaciones_tel_fecha
ON notificaciones (telefono, fecha_envio);

CREATE INDEX IF NOT EXISTS idx_notificaciones_estado
ON notificaciones (estado);

CREATE INDEX IF NOT EXISTS idx_notificaciones_fecha
ON notificaciones (fecha_envio);

CREATE TABLE IF NOT EXISTS seguimiento_clientes (
    telefono      TEXT NOT NULL,
    placa         TEXT DEFAULT '',
    ultimo_ciclo  INTEGER DEFAULT 0,
    updated_at    TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (telefono, placa)
);
`
var db *sql.DB

func GetConnection() *sql.DB {
	db, err := sql.Open("sqlite", "./notiflow.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")

	if err != nil {
		logger.L.Sugar().Errorf("Failed to connect to database %v", err)
		panic(err)
	}

	if err := db.Ping(); err != nil {
		logger.L.Sugar().Errorf("Failed to ping database %v", err)
		panic(err)
	}

	logger.L.Sugar().Infof("Successfully connected to database")

	return db
}

func Init(d *sql.DB) {
	db = d
	_, err := db.Exec(schema)

	if err != nil {
		logger.L.Sugar().Errorf("Failed to create tables %v", err)
		panic(err)
	}
	logger.L.Info("Database initialized successfully with default schema")
}

func WasRecentlyNotified(phone string, placa string, days int) (bool, error) {
	limit := time.Now().AddDate(0, 0, -days)
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM notificaciones
        WHERE telefono = ?
          AND placa = ?
          AND estado = 'enviado'
          AND fecha_envio > ?
    `, phone, placa, limit.Format("2006-01-02 15:04:05")).Scan(&count)

	if err != nil {
		logger.L.Sugar().Errorf("Error checking notification for %s: %v", phone, err)
		return false, err
	}

	return count > 0, nil
}

func GetUltimoCiclo(phone string, placa string) (int, error) {
	var ciclo int
	err := db.QueryRow(`
		SELECT ultimo_ciclo FROM seguimiento_clientes
		WHERE telefono = ? AND placa = ?
	`, phone, placa).Scan(&ciclo)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		logger.L.Sugar().Errorf("Error obteniendo ultimo ciclo para %s/%s: %v", phone, placa, err)
		return 0, err
	}
	return ciclo, nil
}

func UpdateUltimoCiclo(phone string, placa string, ciclo int) error {
	_, err := db.Exec(`
		INSERT INTO seguimiento_clientes (telefono, placa, ultimo_ciclo, updated_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(telefono, placa) DO UPDATE SET
			ultimo_ciclo = excluded.ultimo_ciclo,
			updated_at = datetime('now')
	`, phone, placa, ciclo)
	if err != nil {
		logger.L.Sugar().Errorf("Error actualizando ciclo para %s/%s: %v", phone, placa, err)
		return err
	}
	return nil
}

func WasNotifiedToday(phone string) (bool, error) {
	today := time.Now().Format("2006-01-02")

	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notificaciones
		WHERE telefono = ?
		  AND estado = 'enviado'
		  AND date(fecha_envio) = ?
	`, phone, today).Scan(&count)

	if err != nil {
		logger.L.Sugar().Errorf("Error checking today notification for %s: %v", phone, err)
		return false, err
	}

	return count > 0, nil
}

func CreateNotification(
	phone string,
	placa string,
	name string,
	description string,
	value float64,
	daysOverdue int,
	message string,
	status string,
	errorDetail *string,
) error {

	_, err := db.Exec(`
		INSERT INTO notificaciones
		(telefono, placa,  nombre, descripcion, valor, dias_vencidos, mensaje_enviado, estado, error_detalle)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		phone,
		placa,
		name,
		description,
		value,
		daysOverdue,
		message,
		status,
		errorDetail,
	)

	if err != nil {
		logger.L.Sugar().Errorf("Error inserting notification: %v", err)
		return err
	}

	return nil
}
