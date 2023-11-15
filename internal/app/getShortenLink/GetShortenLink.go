package getshortenlink

import (
	storage "github.com/knstch/shortener/internal/app/storage"
)

// Функция принимает URL строку, ищет совпадение в хранилище
// по значению и выдает ключ-длинную ссылку
func GetShortenLink(url string, URLstorage storage.MemStorage) (string, error) {
	return URLstorage.FindLink(url)
}
