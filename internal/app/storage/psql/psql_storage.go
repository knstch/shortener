package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	common "github.com/knstch/shortener/internal/app/common"
	logger "github.com/knstch/shortener/internal/app/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// PsqURLlStorage - сущность, хранящая соединение с БД.
type PsqURLlStorage struct {
	db *sql.DB
}

// URLs используется для кодирования данных в JSON формат.
type URLs struct {
	LongLink  string `json:"original_url"`
	ShortLink string `json:"short_url"`
}

// NewPsqlStorage возвращает соединение с БД.
func NewPsqlStorage(db *sql.DB) *PsqURLlStorage {
	return &PsqURLlStorage{db: db}
}

func (storage *PsqURLlStorage) findShortLink(ctx context.Context, longLink string) string {

	var link string

	db := bun.NewDB(storage.db, pgdialect.New())

	err := db.NewSelect().
		TableExpr("shorten_URLs").
		Column("short_link").
		Where("long_link = ?", longLink).
		Scan(ctx, &link)
	if err != nil {
		logger.ErrorLogger("Error scanning data: ", err)
	}

	return link
}

func (storage *PsqURLlStorage) insertData(ctx context.Context, longLink string, UserID int) (string, error) {

	generatedShortLink := common.ShortLinkGenerator(5)

	type ShortenUrls struct {
		ShortLink string `bun:"short_link"`
		LongLink  string `bun:"long_link"`
		UserID    int    `bun:"user_id"`
		Deleted   bool   `bun:"deleted"`
	}

	link := &ShortenUrls{
		ShortLink: generatedShortLink,
		LongLink:  longLink,
		UserID:    UserID,
		Deleted:   false,
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	_, err := db.NewInsert().
		Model(link).
		Exec(ctx)
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		shortLink := storage.findShortLink(ctx, longLink)
		return shortLink, err
	}

	return generatedShortLink, nil
}

// PostLink записывает длинную ссылку в хранилище и отдает короткую ссылку и ошибку.
func (storage *PsqURLlStorage) PostLink(ctx context.Context, longLink string, URLaddr string, UserID int) (string, error) {
	shortenLink, err := storage.insertData(ctx, longLink, UserID)
	if err != nil {
		logger.ErrorLogger("Have an error inserting data in PostLink: ", err)
		return URLaddr + "/" + shortenLink, err
	}
	return URLaddr + "/" + shortenLink, nil
}

// FindLink ищет ссылку по короткому адресу и отдает длинную ссылку.
func (storage *PsqURLlStorage) FindLink(ctx context.Context, url string) (string, bool, error) {
	var longLink struct {
		URL          string `bun:"long_link"`
		DeleteStatus bool   `bun:"deleted"`
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	err := db.NewSelect().
		TableExpr("shorten_URLs").
		Model(&longLink).
		Where("short_link = ?", url).
		Scan(ctx)

	if err != nil {
		logger.ErrorLogger("Can't write longLink: ", err)
		return "", longLink.DeleteStatus, err
	}
	return longLink.URL, longLink.DeleteStatus, nil
}

// GetURLsByID получает ID клиента из куки и возвращает все ссылки отправленные им.
func (storage *PsqURLlStorage) GetURLsByID(ctx context.Context, id int, URLaddr string) ([]byte, error) {

	var userIDs []URLs

	var bunURLS struct {
		LongLink  string `bun:"long_link"`
		ShortLink string `bun:"short_link"`
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	rows, err := db.NewSelect().
		TableExpr("shorten_URLs").
		Model(&bunURLS).
		Where("user_id = ?", id).
		Rows(ctx)
	if err != nil {
		logger.ErrorLogger("Error getting data: ", err)
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		logger.ErrorLogger("Have errors readong rows: ", err)
		return nil, err
	}

	for rows.Next() {
		var links URLs
		err = rows.Scan(&links.LongLink, &links.ShortLink)
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
	err = rows.Close()
	if err != nil {
		logger.ErrorLogger("Can't close rows: ", err)
	}
	return jsonUserIDs, nil
}

// DeleteURLs удаляет ссылки, отправленные клиентом при том условии, что он их загрузил.
func (storage *PsqURLlStorage) DeleteURLs(ctx context.Context, id int, shortURLs []string) error {
	inputCh := deleteURLsGenerator(ctx, shortURLs)

	storage.bulkDeleteStatusUpdate(id, inputCh)

	return nil
}

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

func (storage *PsqURLlStorage) bulkDeleteStatusUpdate(id int, inputChs ...chan string) {
	var wg sync.WaitGroup

	deleteUpdate := func(c chan string) {
		var linksToDelete []string
		for shortenLink := range c {
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

	wg.Add(len(inputChs))

	for _, c := range inputChs {
		go deleteUpdate(c)
	}
	wg.Wait()
}

func (storage *PsqURLlStorage) GetStats(ctx context.Context) ([]byte, error) {
	urlsStatRaw := storage.db.QueryRowContext(ctx, "SELECT COUNT (*) as urlsCount from shorten_URLs")
	userStatRaw := storage.db.QueryRowContext(ctx, "SELECT COUNT (DISTINCT user_id) as uniqueUsers from shorten_URLs")

	var urlsStat, usersStat int

	if err := urlsStatRaw.Scan(&urlsStat); err != nil {
		return nil, err
	}

	if err := userStatRaw.Scan(&usersStat); err != nil {
		return nil, err
	}

	readyStats, err := json.Marshal(common.Stats{
		URLs:  urlsStat,
		Users: usersStat,
	})
	if err != nil {
		return nil, err
	}

	return readyStats, nil
}
