package insertdata

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/knstch/shortener/internal/app/logger"
)

// Запись данных в БД
func InsertData(dsn string, shortLink string, longLink string, db *sql.DB) error {

	tx, err := db.Begin()
	if err != nil {
		logger.ErrorLogger("Can't make a transaction: ", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	preparedRequest, err := db.PrepareContext(ctx, "INSERT INTO shorten_URLs(short_link, long_link) VALUES ($1, $2);")
	if err != nil {
		logger.ErrorLogger("Can't prepare request: ", err)
		return err
	}
	defer preparedRequest.Close()

	_, err = preparedRequest.ExecContext(ctx, shortLink, longLink)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
