package dbconnect

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	errorLogger "github.com/knstch/shortener/internal/app/logger"
)

type DBConnection struct {
	db *sql.DB
}

func NewDBConnection(db *sql.DB) *DBConnection {
	return &DBConnection{db: db}
}

// Пинг базы данных
func (db *DBConnection) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.db.PingContext(ctx); err != nil {
		errorLogger.ErrorLogger("Error pinging DB: ", err)
		return err
	}
	return nil
}
