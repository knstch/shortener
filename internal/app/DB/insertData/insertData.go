package insertdata

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	checkDuplicate "github.com/knstch/shortener/internal/app/DB/checkDuplicate"
	logger "github.com/knstch/shortener/internal/app/logger"
)

// Генератор короткой ссылки
func shortLinkGenerator(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// Запись данных в БД
func InsertData(longLink string, db *sql.DB) (string, int, error) {

	generatedShortLink := shortLinkGenerator(5)

	var statusCode int

	tx, err := db.Begin()
	if err != nil {
		logger.ErrorLogger("Can't make a transaction: ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	preparedRequest, err := db.PrepareContext(ctx, "INSERT INTO shorten_URLs(short_link, long_link) VALUES ($1, $2);")
	if err != nil {
		logger.ErrorLogger("Can't prepare request: ", err)
		return "", 500, err
	}
	defer preparedRequest.Close()

	var pgErr *pgconn.PgError

	_, err = preparedRequest.ExecContext(ctx, generatedShortLink, longLink)
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		shortLink := checkDuplicate.FindShortLink(longLink, db)
		statusCode = 409
		tx.Rollback()
		return shortLink, statusCode, nil
	}
	statusCode = 201

	createIndex, err := db.PrepareContext(ctx, "CREATE INDEX long_link ON shorten_urls (long_link)")
	if err != nil {
		logger.ErrorLogger("Can't create indexes: ", err)
	}
	_, err = createIndex.ExecContext(ctx)
	if err != nil {
		tx.Rollback()
	}

	tx.Commit()

	return generatedShortLink, statusCode, nil
}
