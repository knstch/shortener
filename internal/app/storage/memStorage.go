package storage

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	config "github.com/knstch/shortener/cmd/config"
)

type MemStorage struct {
	Data    map[string]string `json:"links"`
	Counter int               `json:"counter"`
	Mu      *sync.Mutex       `json:"-"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Mu:   &sync.Mutex{},
		Data: make(map[string]string),
	}
}

// Сохраняем данные в файл
func (storage *MemStorage) Save(fname string) error {
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
func (storage *MemStorage) Load(fname string) error {
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
func (storage MemStorage) FindLink(url string) (string, error) {
	value, ok := storage.Data[url]
	if !ok {
		return "", nil
	}
	return value, nil
}

// Запись ссылки в базу данных, json хранилище или in-memory. Если идет запись дубликата в БД,
// возвращается уже существующая ссылка
func (storage *MemStorage) PostLink(longLink string, URLaddr string) (string, int) {
	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	storage.Counter++
	storage.Data["shortenLink"+strconv.Itoa(storage.Counter)] = longLink
	storage.Save(config.ReadyConfig.FileStorage)
	return URLaddr + "/shortenLink" + strconv.Itoa(storage.Counter), 201
}
