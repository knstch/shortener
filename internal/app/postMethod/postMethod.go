package postmethod

import "strconv"

var counter int

type Storage struct {
	Data map[string]string
}

// Map-хранилище
var StorageURLs = Storage{
	Data: make(map[string]string),
}

// Принимает запрос в формате string и указатель на хранилище,
// генерит короткую ссылку и записывает данные в формате ключ-значение
func PostMethod(reqBody string, URLstorage *Storage, URLaddr string) string {
	counter++
	URLstorage.Data[reqBody] = "shortenLink" + strconv.Itoa(counter)
	shortenLink := URLaddr + "/" + URLstorage.Data[string(reqBody)]
	return shortenLink
}
