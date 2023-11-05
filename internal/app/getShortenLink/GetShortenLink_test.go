package getshortenlink

import (
	"sync"
	"testing"

	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
	"github.com/stretchr/testify/assert"
)

func TestGetShortenLink(t *testing.T) {

	var testStorage = URLstorage.Storage{
		Data: make(map[string]string),
		Mu:   &sync.Mutex{},
	}
	testStorage.Data["shortenLink1"] = "https://practicum.yandex.ru/"
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
			longLink, err := GetShortenLink((tests[i].request), testStorage)
			assert.NoError(t, err)
			assert.Equal(t, longLink, tests[i].want)
		})
	}
}
