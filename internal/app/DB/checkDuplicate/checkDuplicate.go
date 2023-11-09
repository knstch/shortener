package checkduplicate

import (
	"context"
	"database/sql"

	"github.com/knstch/shortener/internal/app/logger"
)

// Ищет дубликат по длинной ссылке
func CheckDuplicate(dsn string, longLink string, db *sql.DB) bool {
	var ifShortLinkExists bool

	checkLink := db.QueryRowContext(context.Background(), "SELECT EXISTS (SELECT short_link FROM "+
		"shorten_URLs WHERE long_link = $1)", longLink)
	err := checkLink.Scan(&ifShortLinkExists)
	if err != nil {
		logger.ErrorLogger("Error scanning data: ", err)
		return false
	}
	if !ifShortLinkExists {
		return false
	}
	return true
}

// Ищет короткую ссылку по длинной ссылке
func FindShortLink(dsn string, longLink string, db *sql.DB) string {
	var shortLink string

	row := db.QueryRowContext(context.Background(), "SELECT short_link from shorten_URLs WHERE long_link = $1", longLink)
	err := row.Scan(&shortLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return ""
	}
	return shortLink
}
