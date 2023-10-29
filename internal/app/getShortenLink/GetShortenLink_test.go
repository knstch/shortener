package getshortenlink

import (
	"sync"
	"testing"

	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
	"github.com/stretchr/testify/assert"
)

func TestGetShortenLink(t *testing.T) {

	var testStorage = URLstorage.Storage{
		Data: []URLstorage.Links{},
		Mu:   &sync.Mutex{},
	}
	var data = URLstorage.Links{
		ShortLink: "shortenLink1",
		LongLink:  "https://practicum.yandex.ru/",
	}
	testStorage.Data = append(testStorage.Data, data)
	tests := []struct {
		name    string
		want    string
		request string
	}{
		{
			name:    "First test",
			want:    "https://practicum.yandex.ru/",
			request: "shortenLink1",
		},
		{
			name:    "Bad request test",
			want:    "",
			request: "asdwasd",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, GetShortenLink((tests[i].request), testStorage), tests[i].want)
		})
	}
}
