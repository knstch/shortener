package urlstorage

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"sync"

	config "github.com/knstch/shortener/cmd/config"
	checkDuplicate "github.com/knstch/shortener/internal/app/DB/checkDuplicate"
	findData "github.com/knstch/shortener/internal/app/DB/findData"
	insertData "github.com/knstch/shortener/internal/app/DB/insertData"
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

// Генератор короткой ссылки
func shortLinkGenerator(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
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
		return findData.FindData(config.ReadyConfig.DSN, url)
	} else {
		value, ok := storage.Data[url]
		if !ok {
			return "", nil
		}
		return value, nil
	}
}

// Запись ссылки в базу данных или json хранилище. Если идет запись дубликата в БД,
// возвращается уже существующая ссылка
func (storage *Storage) PostLink(reqBody string, URLaddr string) string {
	if config.ReadyConfig.DSN != "" {
		if !checkDuplicate.CheckDuplicate(config.ReadyConfig.DSN, reqBody) {
			shortenLink := shortLinkGenerator(5)
			insertData.InsertData(config.ReadyConfig.DSN, shortenLink, reqBody)
			return URLaddr + "/" + shortenLink
		}
		return URLaddr + "/" + checkDuplicate.FindShortLink(config.ReadyConfig.DSN, reqBody)
	} else {
		storage.Mu.Lock()
		defer storage.Mu.Unlock()
		storage.Counter++
		storage.Data["shortenLink"+strconv.Itoa(storage.Counter)] = reqBody
		storage.Save(config.ReadyConfig.FileStorage)
		return URLaddr + "/shortenLink" + strconv.Itoa(storage.Counter)
	}
}
