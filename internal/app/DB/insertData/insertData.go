package insertdata

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/knstch/shortener/internal/app/logger"
)

func InsertData(dsn string, shortLink string, longLink string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorLogger("Can't open a new database: ", err)
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	writeData, err := db.ExecContext(ctx, "INSERT INTO shorten_URLs(short_link, long_link) VALUES ($1, $2);", longLink, shortLink)
	if err != nil {
		logger.ErrorLogger("Error writing data: ", err)
	}
	rows, err := writeData.RowsAffected()
	if err != nil {
		logger.ErrorLogger("Error when insert data: ", err)
	}

	logger.InfoLogger("Rows affected when inserting data: %d" + string(rune(rows)))
	return nil
}
