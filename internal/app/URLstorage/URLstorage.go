package urlstorage

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/knstch/shortener/cmd/config"
)

type (
	Storage struct {
		Data    []Links     `json:"links"`
		Counter int         `json:"counter"`
		Mu      *sync.Mutex `json:"-"`
	}
	Links struct {
		ShortLink string `json:"short_link"`
		LongLink  string `json:"long_link"`
	}
)

var StorageURLs = Storage{
	Mu: &sync.Mutex{},
}

// Сохраняем данные в файл
func (storage *Storage) Save(fname string) error {
	data, err := json.Marshal(storage)
	if err != nil {
		return err
	}
	return os.WriteFile(fname, data, 0666)
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
	for _, links := range storage.Data {
		if links.ShortLink == url {
			return links.LongLink
		}
	}
	return ""
}

func (storage *Storage) PostLink(reqBody string, URLaddr string) string {
	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	storage.Counter++
	storage.Data = append(storage.Data, Links{
		ShortLink: "shortenLink" + strconv.Itoa(storage.Counter),
		LongLink:  reqBody,
	})
	shortenLink := URLaddr + "/" + "shortenLink" + strconv.Itoa(storage.Counter)
	storage.Save(config.ReadyConfig.FileStorage)
	return shortenLink
}
