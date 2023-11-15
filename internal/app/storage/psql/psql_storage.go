package psql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/knstch/shortener/internal/app/logger"
)

type PsqlStorage struct {
	db *sql.DB
}

func NewPsqlStorage(db *sql.DB) *PsqlStorage {
	return &PsqlStorage{db: db}
}

func (storage *PsqlStorage) FindLink(url string) (string, error) {
	return storage.FindData(url)
}

func (storage *PsqlStorage) PostLink(ctx context.Context, longLink string, URLaddr string) (string, error) {
	shortenLink, err := storage.insertData(ctx, longLink)
	if err != nil {
		logger.ErrorLogger("Have an error inserting data in PostLink: ", err)
		return "", err
	}
	return URLaddr + "/" + shortenLink, nil
}

func (storage *PsqlStorage) insertData(ctx context.Context, longLink string) (string, error) {
	generatedShortLink := shortLinkGenerator(5)

	tx, err := storage.db.Begin()
	if err != nil {
		logger.ErrorLogger("Can't make a transaction: ", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	preparedRequest, err := storage.db.PrepareContext(ctx, "INSERT INTO shorten_URLs(short_link, long_link) VALUES ($1, $2);")
	if err != nil {
		logger.ErrorLogger("Can't prepare request: ", err)
		return "", err
	}
	defer preparedRequest.Close()

	var pgErr *pgconn.PgError

	_, err = preparedRequest.ExecContext(ctx, generatedShortLink, longLink)
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		shortLink := storage.FindShortLink(longLink)
		tx.Rollback()
		return shortLink, nil
	}

	createIndex, err := storage.db.PrepareContext(ctx, "CREATE INDEX long_link ON shorten_urls (long_link)")
	if err != nil {
		logger.ErrorLogger("Can't create indexes: ", err)
	}
	_, err = createIndex.ExecContext(ctx)
	if err != nil {
		tx.Rollback()
	}

	tx.Commit()

	return generatedShortLink, nil
}

// Поиск длинной ссылки по короткой
func (storage *PsqlStorage) FindData(shortLink string) (string, error) {
	var longLink string

	row := storage.db.QueryRowContext(context.Background(), "SELECT long_link from shorten_URLs WHERE short_link = $1", shortLink)
	err := row.Scan(&longLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return "", err
	}
	return longLink, nil
}

// Ищет короткую ссылку по длинной ссылке
func (storage *PsqlStorage) FindShortLink(longLink string) string {
	var shortLink string

	row := storage.db.QueryRowContext(context.Background(), "SELECT short_link from shorten_URLs WHERE long_link = $1", longLink)
	err := row.Scan(&shortLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return ""
	}
	return shortLink
}
