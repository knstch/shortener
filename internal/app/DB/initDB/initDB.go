package initdb

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/knstch/shortener/internal/app/errorLogger"
)

func InitDB(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorLogger("Can't open a new database: ", err)
		return err
	}
	defer db.Close()

	initialization := `CREATE TABLE IF NOT EXISTS shorten_URLs(long_link text, short_link text)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createDB, err := db.ExecContext(ctx, initialization)
	if err != nil {
		logger.ErrorLogger("Can't init database: ", err)
		return err
	}

	rows, err := createDB.RowsAffected()
	if err != nil {
		logger.ErrorLogger("Error when getting rows affected: ", err)
		return err
	}
	logger.InfoLogger("Rows affected when creating table: " + string(rune(rows)))

	return nil
}
