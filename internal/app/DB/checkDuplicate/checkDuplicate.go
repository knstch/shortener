package checkduplicate

import (
	"context"
	"database/sql"

	"github.com/knstch/shortener/internal/app/logger"
)

// Ищет короткую ссылку по длинной ссылке
func FindShortLink(longLink string, db *sql.DB) string {
	var shortLink string

	row := db.QueryRowContext(context.Background(), "SELECT short_link from shorten_URLs WHERE long_link = $1", longLink)
	err := row.Scan(&shortLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return ""
	}
	return shortLink
}
