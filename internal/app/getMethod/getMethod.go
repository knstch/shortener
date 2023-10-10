package getmethod

import (
	postMethod "github.com/knstch/shortener/internal/app/postMethod"
)

// Функция принимает URL строку, ищет совпадение в хранилище
// по значению и выдает ключ-длинную ссылку
func GetMethod(url string, URLstorage postMethod.Storage) string {
	for k, v := range URLstorage.Data {
		if v == url {
			return k
		}
	}
	return ""
}
