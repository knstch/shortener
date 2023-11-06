package finddata

import (
	"context"
	"database/sql"

	"github.com/knstch/shortener/internal/app/logger"
)

// Поиск длинной ссылки по короткой
func FindData(dsn string, shortLink string) (string, error) {
	var longLink string
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorLogger("Can't open a new database: ", err)
		return "", err
	}
	defer db.Close()

	row := db.QueryRowContext(context.Background(), "SELECT long_link from shorten_URLs WHERE short_link = $1", shortLink)
	err = row.Scan(&longLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return "", err
	}
	return longLink, nil
}
