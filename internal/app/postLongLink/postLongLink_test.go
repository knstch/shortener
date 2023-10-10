package postlonglink

import (
	"sync"
	"testing"

	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
	"github.com/stretchr/testify/assert"
)

func TestPostLongLink(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		request string
	}{
		{
			name:    "First test",
			want:    "http://localhost:8080/shortenLink1",
			request: "https://practicum.yandex.ru/",
		},
		{
			name:    "Second test",
			want:    "http://localhost:8080/shortenLink2",
			request: "https://practicum.yandex.ru/2",
		},
	}
	var testStorage = URLstorage.Storage{
		Data: make(map[string]string),
		Mu:   &sync.Mutex{},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, PostLongLink(tests[i].request, &testStorage, "http://localhost:8080"), tests[i].want)
		})
	}
}
