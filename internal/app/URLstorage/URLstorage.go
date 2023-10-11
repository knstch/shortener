package urlstorage

import (
	"strconv"
	"sync"
)

type Storage struct {
	Data    map[string]string
	Counter int
	Mu      *sync.Mutex
}

var StorageURLs = Storage{
	Data: make(map[string]string),
	Mu:   &sync.Mutex{},
}

func (s Storage) FindLink(url string) string {
	value, ok := s.Data[url]
	if !ok {
		return ""
	}
	return value
}

func (s *Storage) PostLink(reqBody string, URLaddr string) string {
	s.Counter++
	s.Data["shortenLink"+strconv.Itoa(s.Counter)] = reqBody
	shortenLink := URLaddr + "/" + "shortenLink" + strconv.Itoa(s.Counter)
	return shortenLink
}
