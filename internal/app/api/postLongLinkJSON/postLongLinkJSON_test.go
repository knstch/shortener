package postlonglinkjson

import (
	"testing"

	config "github.com/knstch/shortener/cmd/config"
	"github.com/stretchr/testify/assert"
)

func TestPostlonglinkjson(t *testing.T) {
	config.ParseConfig()
	tests := []struct {
		name    string
		want    string
		request string
	}{
		{
			name: "First test",
			want: `{
				"result": "http://localhost:8080/shortenLink1"
			}`,
			request: `{
				"url": "http://example.com/"
			}`,
		},
		{
			name: "Second test",
			want: `{
				"result": "http://localhost:8080/shortenLink2"
			}`,
			request: `{
				"url": "https://practicum.yandex.ru/"
			}`,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortLink, _ := PostLongLinkJSON([]byte(tests[i].request))
			assert.JSONEq(t, string(shortLink), tests[i].want)
		})
	}
}
