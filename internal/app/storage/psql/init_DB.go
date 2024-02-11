package psql

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/knstch/shortener/internal/app/logger"
)

// Инициализация таблицы shorten_URLs с полями long_link text и short_link text
func InitDB(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	initialization := `CREATE TABLE IF NOT EXISTS shorten_URLs(
		 long_link varchar(255) UNIQUE,
		 short_link varchar(255), 
		 correlation_id varchar(255),
		 user_id INT,
		 deleted BOOLEAN);`

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
