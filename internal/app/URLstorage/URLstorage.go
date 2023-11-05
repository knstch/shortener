package urlstorage

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	config "github.com/knstch/shortener/cmd/config"
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
func (storage Storage) FindLink(url string) string {
	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	value, ok := storage.Data[url]
	if !ok {
		return ""
	}
	return value
}

func (storage *Storage) PostLink(reqBody string, URLaddr string) string {
	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	storage.Counter++
	storage.Data["shortenLink"+strconv.Itoa(storage.Counter)] = reqBody
	shortenLink := URLaddr + "/" + "shortenLink" + strconv.Itoa(storage.Counter)
	storage.Save(config.ReadyConfig.FileStorage)
	defer insertData.InsertData(config.ReadyConfig.DSN, "shortenLink"+strconv.Itoa(storage.Counter), reqBody)
	return shortenLink
}
