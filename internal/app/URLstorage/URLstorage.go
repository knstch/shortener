package urlstorage

import "sync"

type Storage struct {
	Data    map[string]string
	Counter int
	Mu      *sync.Mutex
}

var StorageURLs = Storage{
	Data: make(map[string]string),
	Mu:   &sync.Mutex{},
}
