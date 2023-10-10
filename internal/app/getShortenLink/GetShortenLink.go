package getshortenlink

import (
	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
)

// Функция принимает URL строку, ищет совпадение в хранилище
// по значению и выдает ключ-длинную ссылку
func GetShortenLink(url string, URLstorage URLstorage.Storage) string {
	URLstorage.Mu.Lock()
	defer URLstorage.Mu.Unlock()
	value, ok := URLstorage.Data[url]
	if !ok {
		return ""
	}
	return value
}
