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
func PostMethod(reqBody string, URLstorage *Storage) string {
	counter++
	URLstorage.Data[reqBody] = "shortenLink" + strconv.Itoa(counter)
	shortenLink := "http://localhost:8080/" + URLstorage.Data[string(reqBody)]
	return shortenLink
}
