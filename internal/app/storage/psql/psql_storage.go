package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	logger "github.com/knstch/shortener/internal/app/logger"
)

type PsqURLlStorage struct {
	db *sql.DB
}

func NewPsqlStorage(db *sql.DB) *PsqURLlStorage {
	return &PsqURLlStorage{db: db}
}

// Ищет короткую ссылку по длинной ссылке
func (storage *PsqURLlStorage) findShortLink(longLink string) string {
	var shortLink string

	row := storage.db.QueryRowContext(context.Background(), "SELECT short_link from shorten_URLs WHERE long_link = $1", longLink)
	err := row.Scan(&shortLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return ""
	}
	return shortLink
}

// Запись данных в БД
func (storage *PsqURLlStorage) insertData(ctx context.Context, longLink string, UserID int) (string, error) {

	generatedShortLink := shortLinkGenerator(5)

	tx, err := storage.db.Begin()
	if err != nil {
		logger.ErrorLogger("Can't make a transaction: ", err)
		return "", err
	}

	preparedRequest, err := storage.db.PrepareContext(ctx, "INSERT INTO shorten_URLs(short_link, long_link, user_id) VALUES ($1, $2, $3);")
	if err != nil {
		logger.ErrorLogger("Can't prepare request: ", err)
		return "", err
	}
	defer preparedRequest.Close()

	var pgErr *pgconn.PgError

	_, err = preparedRequest.ExecContext(ctx, generatedShortLink, longLink, UserID)
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		shortLink := storage.findShortLink(longLink)
		tx.Rollback()
		return shortLink, err
	}

	tx.Commit()

	return generatedShortLink, nil
}

func (storage *PsqURLlStorage) PostLink(ctx context.Context, longLink string, URLaddr string, UserID int) (string, error) {
	shortenLink, err := storage.insertData(ctx, longLink, UserID)
	if err != nil {
		logger.ErrorLogger("Have an error inserting data in PostLink: ", err)
		return URLaddr + "/" + shortenLink, err
	}
	return URLaddr + "/" + shortenLink, nil
}

// Поиск длинной ссылки по короткой
func (storage *PsqURLlStorage) FindLink(url string) (string, error) {
	var longLink string

	row := storage.db.QueryRowContext(context.Background(), "SELECT long_link from shorten_URLs WHERE short_link = $1", url)
	err := row.Scan(&longLink)
	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return "", err
	}
	return longLink, nil
}

type URLs struct {
	LongLink  string `json:"original_url"`
	ShortLink string `json:"short_url"`
}

func (storage *PsqURLlStorage) GetURLsByID(ctx context.Context, id int, URLaddr string) ([]byte, error) {

	var userIDs []URLs

	allIDs, err := storage.db.QueryContext(ctx, "SELECT long_link, short_link from shorten_URLs WHERE user_id = $1;", id)
	if err != nil {
		logger.ErrorLogger("Error getting batch data: ", err)
		return nil, err
	}
	defer allIDs.Close()

	for allIDs.Next() {
		var links URLs
		err := allIDs.Scan(&links.LongLink, &links.ShortLink)
		if err != nil {
			logger.ErrorLogger("Error scanning data: ", err)
			return nil, err
		}
		userIDs = append(userIDs, URLs{
			LongLink:  links.LongLink,
			ShortLink: URLaddr + "/" + links.ShortLink,
		})
	}
	jsonUserIDs, err := json.Marshal(userIDs)
	if err != nil {
		logger.ErrorLogger("Can't marshal IDs: ", err)
		return nil, err
	}

	return jsonUserIDs, nil
}
