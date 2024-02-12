// Модуль storage имеет методы для завимодействия с memstorage.
package storage

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	config "github.com/knstch/shortener/cmd/config"
)

// MemStorage хранит ссылки в виде мапы, счетчик для сокращения ссылок и mutex.
type MemStorage struct {
	Data    map[string]string `json:"links"`
	Counter int               `json:"counter"`
	Mu      *sync.Mutex       `json:"-"`
}

// NewMemStorage возвращает новое in-memory хранилище. 
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Mu:   &sync.Mutex{},
		Data: make(map[string]string),
	}
}

func (storage *MemStorage) save(fname string) error {
	data, err := json.MarshalIndent(storage, "", "   ")
	if err != nil {
		return err
	}
	if len(storage.Data)%30 == 0 || len(storage.Data) < 2 {
		return os.WriteFile(fname, data, 0666)
	}
	return nil
}

func (storage *MemStorage) load(fname string) error {
	data, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, storage); err != nil {
		return err
	}
	return nil
}

// FindLink ищет ссылку по короткому адресу и отдает длинную ссылку.
func (storage MemStorage) FindLink(url string) (string, bool, error) {
	storage.load(config.ReadyConfig.FileStorage)
	value, ok := storage.Data[url]
	if !ok {
		return "", false, nil
	}
	return value, false, nil
}

// IntegrityError возвращает тип ошибки IntegrityError.
type IntegrityError struct {
	msg string
}

func (e *IntegrityError) Error() string {
	return e.msg
}

func NewIntegrityError(msg string) error {
	return &IntegrityError{msg: msg}
}

// PostLink записывает длинную ссылку в хранилище и отдает короткую ссылку и ошибку.
func (storage *MemStorage) PostLink(_ context.Context, longLink string, URLaddr string, _ int) (string, error) {
	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	for k, v := range storage.Data {
		if v == longLink {
			return URLaddr + "/" + k, NewIntegrityError("Duplicate")
		}
	}
	storage.Counter++
	storage.Data["shortenLink"+strconv.Itoa(storage.Counter)] = longLink
	storage.save(config.ReadyConfig.FileStorage)
	return URLaddr + "/shortenLink" + strconv.Itoa(storage.Counter), nil
}

func (storage *MemStorage) GetURLsByID(ctx context.Context, id int, URLaddr string) ([]byte, error) {
	return []byte("Memory storage can't operate with user IDs"), nil
}

func (storage *MemStorage) DeleteURLs(ctx context.Context, id int, shortURLs []string) error {
	return nil
}
