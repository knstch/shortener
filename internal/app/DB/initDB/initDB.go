package initdb

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	config "github.com/knstch/shortener/cmd/config"
	logger "github.com/knstch/shortener/internal/app/logger"
)

// Инициализация таблицы shorten_URLs с полями long_link text и short_link text
func InitDB(dsn string) error {
	db, err := sql.Open("pgx", config.ReadyConfig.DSN)
	if err != nil {
		logger.ErrorLogger("Can't open connection: ", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	initialization := `CREATE TABLE IF NOT EXISTS shorten_URLs(long_link text, short_link text);`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, initialization)
	if err != nil {
		tx.Rollback()
		return err
	}

	logger.InfoLogger("Table inited")

	return tx.Commit()
}
