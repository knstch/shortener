package urlstorage

import (
	"database/sql"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	config "github.com/knstch/shortener/cmd/config"
	// checkDuplicate "github.com/knstch/shortener/internal/app/DB/checkDuplicate"
	findData "github.com/knstch/shortener/internal/app/DB/findData"
	insertData "github.com/knstch/shortener/internal/app/DB/insertData"
	logger "github.com/knstch/shortener/internal/app/logger"
)

type (
	Storage struct {
		Data    map[string]string `json:"links"`
		Counter int               `json:"counter"`
		Mu      *sync.Mutex       `json:"-"`
	}
)

var StorageURLs = Storage{
	Mu:   &sync.Mutex{},
	Data: make(map[string]string),
}

// Сохраняем данные в файл
func (storage *Storage) Save(fname string) error {
	data, err := json.MarshalIndent(storage, "", "   ")
	if err != nil {
		return err
	}
	if len(storage.Data)%30 == 0 || len(storage.Data) < 2 {
		return os.WriteFile(fname, data, 0666)
	}
	return nil
}

// Загружаем данные из файла
func (storage *Storage) Load(fname string) error {
	data, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, storage); err != nil {
		return err
	}
	return nil
}

// Ищем ссылку
func (storage Storage) FindLink(url string) (string, error) {
	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		return findData.FindData(config.ReadyConfig.DSN, url, db)
	} else {
		value, ok := storage.Data[url]
		if !ok {
			return "", nil
		}
		return value, nil
	}
}

// Запись ссылки в базу данных, json хранилище или in-memory. Если идет запись дубликата в БД,
// возвращается уже существующая ссылка
func (storage *Storage) PostLink(longLink string, URLaddr string) (string, int) {
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		defer db.Close()
		shortenLink, statusCode, err := insertData.InsertData(longLink, db)
		if err != nil {
			logger.ErrorLogger("Have an error inserting data in PostLink: ", err)
		}
		return URLaddr + "/" + shortenLink, statusCode
		// }

		// return URLaddr + "/" + checkDuplicate.FindShortLink(config.ReadyConfig.DSN, longLink, db), 409
	} else {
		storage.Mu.Lock()
		defer storage.Mu.Unlock()
		storage.Counter++
		storage.Data["shortenLink"+strconv.Itoa(storage.Counter)] = longLink
		storage.Save(config.ReadyConfig.FileStorage)
		return URLaddr + "/shortenLink" + strconv.Itoa(storage.Counter), 201
	}
}
