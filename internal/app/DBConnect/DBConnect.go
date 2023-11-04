package dbconnect

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	errorLogger "github.com/knstch/shortener/internal/app/errorLogger"
)

func OpenConnection(host string, user string, password string, dbname string) error {
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)

	db, err := sql.Open("pgx", ps)
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
