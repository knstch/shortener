package postmethod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostMethod(t *testing.T) {
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
	var testStorage = Storage{
		Data: make(map[string]string),
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, PostMethod(tests[i].request, &testStorage, "http://localhost:8080"), tests[i].want)
		})
	}
}
