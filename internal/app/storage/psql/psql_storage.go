package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	logger "github.com/knstch/shortener/internal/app/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
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

	preparedRequest, err := storage.db.PrepareContext(ctx, "INSERT INTO shorten_URLs(short_link, long_link, user_id, deleted) VALUES ($1, $2, $3, false);")
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
func (storage *PsqURLlStorage) FindLink(url string) (string, bool, error) {
	var longLink struct {
		URL          string `bun:"long_link"`
		DeleteStatus bool   `bun:"deleted"`
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	err := db.NewSelect().
		TableExpr("shorten_URLs").
		Model(&longLink).
		Where("short_link = ?", url).
		Scan(context.Background())

	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return "", longLink.DeleteStatus, err
	}
	return longLink.URL, longLink.DeleteStatus, nil
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
	defer func() {
		_ = allIDs.Close()
		_ = allIDs.Err()
	}()

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

// Удаление URL для пользователей с определенным ID
func (storage *PsqURLlStorage) DeleteURLs(ctx context.Context, id int, shortURLs []string) error {

	context := context.Background()

	inputCh := deleteURLsGenerator(context, shortURLs)

	storage.bulkDeleteStatusUpdate(context, id, inputCh)

	return nil
}

// Генератор канала с ссылками
func deleteURLsGenerator(ctx context.Context, URLs []string) chan string {
	URLsCh := make(chan string)
	go func() {
		defer close(URLsCh)
		for _, data := range URLs {
			select {
			case <-ctx.Done():
				return
			case URLsCh <- data:
			}
		}
	}()
	return URLsCh
}

// Функция-го рутина, принимает канал с ссылками для удаления
func (storage *PsqURLlStorage) deleteUpdate(inputChs chan string, id int) {
	var wg sync.WaitGroup
	var linksToDelete []string
	for shortenLink := range inputChs {
		linksToDelete = append(linksToDelete, shortenLink)
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	_, err := db.NewUpdate().
		TableExpr("shorten_URLs").
		Set("deleted = ?", "true").
		Where("short_link IN (?)", bun.In(linksToDelete)).
		WhereGroup(" AND ", func(uq *bun.UpdateQuery) *bun.UpdateQuery {
			return uq.Where("user_id = ?", id)
		}).
		Exec(context.Background())
	if err != nil {
		logger.ErrorLogger("Can't exec update request: ", err)
	}
	wg.Done()
}

// Удаление всех ссылкой bulk запросом
func (storage *PsqURLlStorage) bulkDeleteStatusUpdate(ctx context.Context, id int, inputChs ...chan string) {
	var wg sync.WaitGroup

	wg.Add(len(inputChs))

	for _, c := range inputChs {
		go storage.deleteUpdate(c, id)
	}

	go func() {
		wg.Wait()
	}()
}
