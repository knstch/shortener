package postlonglink

import (
	storage "github.com/knstch/shortener/internal/app/storage"
)

// Принимает запрос в формате string и указатель на хранилище,
// генерит короткую ссылку и записывает данные в формате ключ-значение
func PostLongLink(reqBody string, URLstorage *URLstorage.MemStorage, URLaddr string) (string, int) {
	return storage.PostLink(reqBody, URLaddr)
}
