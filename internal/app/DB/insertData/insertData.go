package insertdata

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/knstch/shortener/internal/app/logger"
)

// Запись данных в БД
func InsertData(dsn string, shortLink string, longLink string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorLogger("Can't open a new database: ", err)
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db.ExecContext(ctx, "INSERT INTO shorten_URLs(short_link, long_link) VALUES ($1, $2);", shortLink, longLink)

	return nil
}
