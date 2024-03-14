// Модуль dbconnect отвечает за проверку соединения с БД.
package dbconnect

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	errorLogger "github.com/knstch/shortener/internal/app/logger"
)

// DBConnection используется для сохранения sql соеднинеия и взаимодействует с пакетом sql.DB.
type DBConnection struct {
	db *sql.DB
}

// NewDBConnection возвращает сущность DBConnection.
func NewDBConnection(db *sql.DB) *DBConnection {
	return &DBConnection{db: db}
}

// Ping пингует БД.
func (database *DBConnection) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.db.PingContext(ctx); err != nil {
		errorLogger.ErrorLogger("Error pinging DB: ", err)
		return err
	}
	return nil
}
