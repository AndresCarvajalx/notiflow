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
`
var db *sql.DB

func GetConnection() *sql.DB {
	db, err := sql.Open("sqlite", "./notiflow.db")

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

func WasRecentlyNotified(phone string, days int) (bool, error) {
	limit := time.Now().AddDate(0, 0, -days)

	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM notificaciones
		WHERE telefono = ?
		  AND estado = 'enviado'
		  AND fecha_envio > ?
	`, phone, limit.Format("2006-01-02 15:04:05")).Scan(&count)

	if err != nil {
		logger.L.Sugar().Errorf("Error checking notification for %s: %v", phone, err)
		return false, err
	}

	return count > 0, nil
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
		(telefono, nombre, descripcion, valor, dias_vencidos, mensaje_enviado, estado, error_detalle)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		phone,
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
