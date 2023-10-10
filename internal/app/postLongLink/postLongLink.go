package postLongLink

import (
	"strconv"

	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
)

// Принимает запрос в формате string и указатель на хранилище,
// генерит короткую ссылку и записывает данные в формате ключ-значение
func PostLongLink(reqBody string, URLstorage *URLstorage.Storage, URLaddr string) string {
	URLstorage.Mu.Lock()
	URLstorage.Counter++
	URLstorage.Data["shortenLink"+strconv.Itoa(URLstorage.Counter)] = reqBody
	shortenLink := URLaddr + "/" + "shortenLink" + strconv.Itoa(URLstorage.Counter)
	URLstorage.Mu.Unlock()
	return shortenLink
}
