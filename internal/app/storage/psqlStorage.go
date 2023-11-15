package storage

import (
	"database/sql"

	config "github.com/knstch/shortener/cmd/config"
	findData "github.com/knstch/shortener/internal/app/DB/findData"
	insertData "github.com/knstch/shortener/internal/app/DB/insertData"
	logger "github.com/knstch/shortener/internal/app/logger"
)

type PsqlStorage struct {
	db *sql.DB
}

func NewPsqlStorage(db *sql.DB) *PsqlStorage {
	return &PsqlStorage{db: db}
}

func (storage *PsqlStorage) FindLink(url string) (string, error) {
	return findData.FindData(config.ReadyConfig.DSN, url, storage.db)
}

func (storage *PsqlStorage) PostLink(longLink string, URLaddr string) (string, int) {
	shortenLink, statusCode, err := insertData.InsertData(longLink, storage.db)
	if err != nil {
		logger.ErrorLogger("Have an error inserting data in PostLink: ", err)
	}
	return URLaddr + "/" + shortenLink, statusCode
}
