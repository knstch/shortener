package dbconnect

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	errorLogger "github.com/knstch/shortener/internal/app/errorLogger"
)

func OpenConnection(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		errorLogger.ErrorLogger("Can't open a new database: ", err)
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		errorLogger.ErrorLogger("Error pinging DB: ", err)
		return err
	}
	return nil
}
