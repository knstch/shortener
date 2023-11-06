package initdb

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/knstch/shortener/internal/app/logger"
)

// Инициализация таблицы shorten_URLs с полями long_link text и short_link text
func InitDB(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorLogger("Can't open a new database: ", err)
		return err
	}
	defer db.Close()

	initialization := `CREATE TABLE IF NOT EXISTS shorten_URLs(long_link text, short_link text);`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db.ExecContext(ctx, initialization)

	logger.InfoLogger("Table inited")

	return nil
}
