package postlonglink

import (
	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
)

// Принимает запрос в формате string и указатель на хранилище,
// генерит короткую ссылку и записывает данные в формате ключ-значение
func PostLongLink(reqBody string, URLstorage *URLstorage.Storage, URLaddr string) string {
	URLstorage.Mu.Lock()
	defer URLstorage.Mu.Unlock()
	return URLstorage.PostLink(reqBody, URLaddr)
}
