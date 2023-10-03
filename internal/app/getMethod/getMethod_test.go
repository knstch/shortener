package getmethod

import (
	"testing"

	. "github.com/knstch/shortener_url/internal/app/postMethod"
	"github.com/stretchr/testify/assert"
)

func TestGetMethod(t *testing.T) {

	var testStorage = Storage{
		Data: make(map[string]string),
	}
	testStorage.Data["https://practicum.yandex.ru/"] = "shortenLink1"

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
			assert.Equal(t, GetMethod((tests[i].request), testStorage), tests[i].want)
		})
	}
}
