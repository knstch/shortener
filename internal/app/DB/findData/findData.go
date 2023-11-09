package finddata

import (
	"context"
	"database/sql"

	"github.com/knstch/shortener/internal/app/logger"
)

// Поиск длинной ссылки по короткой
func FindData(dsn string, shortLink string, db *sql.DB) (string, error) {
	var longLink string

	row := db.QueryRowContext(context.Background(), "SELECT long_link from shorten_URLs WHERE short_link = $1", shortLink)
	err := row.Scan(&longLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return "", err
	}
	return longLink, nil
}
