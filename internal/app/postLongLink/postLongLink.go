package postlonglink

import (
	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
)

// Принимает запрос в формате string и указатель на хранилище,
// генерит короткую ссылку и записывает данные в формате ключ-значение
func PostLongLink(reqBody string, URLstorage *URLstorage.Storage, URLaddr string) (string, int) {
	shortLink, statusCode := URLstorage.PostLink(reqBody, URLaddr)
	return shortLink, statusCode
}
