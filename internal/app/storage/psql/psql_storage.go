package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/knstch/shortener/cmd/config"
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

type URLToDelete struct {
	shortLinks string
	userID     int
}

func (storage *PsqURLlStorage) DeleteURLs(ctx context.Context, id int, shortURLs []string) error {

	var deleteURLs []URLToDelete

	for _, v := range shortURLs {
		URL := URLToDelete{
			shortLinks: v,
			userID:     id,
		}
		deleteURLs = append(deleteURLs, URL)
	}
	contextFromDeleteURLs := context.Background()
	doneCh := make(chan struct{})
	defer close(doneCh)

	inputCh := deleteURLsGenerator(contextFromDeleteURLs, doneCh, deleteURLs)

	storage.fanOut(contextFromDeleteURLs, doneCh, inputCh)
	// deleteResult := fanIn(doneCh, channels...)

	// for err := range deleteResult {
	// 	if err != nil {
	// 		fmt.Println("Checking errors")
	// 		logger.ErrorLogger("Error deleting URL: ", err)
	// 	}
	// }

	return nil
}

func deleteURLsGenerator(ctx context.Context, doneCh chan struct{}, URLs []URLToDelete) chan URLToDelete {
	URLsCh := make(chan URLToDelete)
	go func() {
		defer close(URLsCh)
		for _, data := range URLs {
			select {
			case <-doneCh:
				return
			case URLsCh <- data:
			}
		}
	}()

	return URLsCh
}

// fanOut принимает канал данных, порождает 100 горутин
func (storage *PsqURLlStorage) fanOut(ctx context.Context, doneCh chan struct{}, inputCh chan URLToDelete) []chan error {
	// количество горутин add
	numWorkers := 10
	// каналы, в которые отправляются результаты
	channels := make([]chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		// получаем канал из горутины add
		go storage.deleteWorker(ctx, doneCh, inputCh)
	}

	// возвращаем слайс каналов
	return channels
}

func (storage *PsqURLlStorage) deleteWorker(ctx context.Context, doneCh chan struct{}, inputCh chan URLToDelete) {
	for link := range inputCh {
		storage.deleteURL(ctx, link)
	}
}

func (storage *PsqURLlStorage) deleteURL(ctx context.Context, URLToDelete URLToDelete) error {
	dbLaunch, err := sql.Open("pgx", config.ReadyConfig.DSN)
	if err != nil {
		logger.ErrorLogger("Can't open connection: ", err)
	}
	defer dbLaunch.Close()
	tx, err := storage.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	db := bun.NewDB(dbLaunch, pgdialect.New())

	_, err = db.NewUpdate().
		TableExpr("shorten_URLs").
		Set("deleted = ?", "true").
		Where("short_link = (?)", URLToDelete.shortLinks).
		WhereGroup(" AND ", func(uq *bun.UpdateQuery) *bun.UpdateQuery {
			return uq.Where("user_id = ?", URLToDelete.userID)
		}).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Can't exec update request: ", err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// func fanIn(doneCh chan struct{}, resultChs ...chan error) chan error {
// 	errorsCh := make(chan error)

// 	var wg sync.WaitGroup

// 	for _, ch := range resultChs {
// 		chCopy := ch
// 		wg.Add(1)

// 		go func() {
// 			defer wg.Done()
// 			fmt.Println("Aboba")
// 			for data := range chCopy {
// 				select {
// 				case <-doneCh:
// 					return
// 				case errorsCh <- data:
// 				}
// 			}
// 		}()
// 	}

// 	go func() {
// 		fmt.Println("Aboba 2")
// 		wg.Wait()
// 		close(errorsCh)
// 	}()
// 	return errorsCh
// }
