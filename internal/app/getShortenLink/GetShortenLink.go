package getshortenlink

import (
	storage "github.com/knstch/shortener/internal/app/URLstorage"
)

// Функция принимает URL строку, ищет совпадение в хранилище
// по значению и выдает ключ-длинную ссылку
func GetShortenLink(url string, URLstorage storage.Storage) string {
	URLstorage.Mu.Lock()
	defer URLstorage.Mu.Unlock()
	return URLstorage.FindLink(url)
}
