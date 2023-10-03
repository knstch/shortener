package postmethod

import "strconv"

var counter int

type Storage struct {
	Data map[string]string
}

var StorageURLs = Storage{
	Data: make(map[string]string),
}

func PostMethod(reqBody string, URLstorage *Storage) string {
	counter++
	URLstorage.Data[reqBody] = "shortenLink" + strconv.Itoa(counter)
	shortenLink := "http://localhost:8080/" + URLstorage.Data[string(reqBody)]
	return shortenLink
}
