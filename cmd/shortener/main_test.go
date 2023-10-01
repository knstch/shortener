package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostMethod(t *testing.T) {
	type Want struct {
		statusCode  int
		contentType string
		body        string
	}
	tests := []struct {
		name    string
		want    Want
		request string
	}{
		{
			name: "First test",
			want: Want{
				statusCode:  201,
				contentType: "text/plain",
				body:        "http://localhost:8080/shortenLink1",
			},
			request: "https://practicum.yandex.ru/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			postMethod(w, request)
			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			decryptedBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(decryptedBody))
		})
	}
}

func TestGetMethod(t *testing.T) {
	type Want struct {
		statusCode  int
		contentType string
		location    string
	}
	tests := []struct {
		name    string
		want    Want
		request string
	}{
		{
			name: "First test",
			want: Want{
				statusCode:  307,
				contentType: "text/plain",
				location:    "https://practicum.yandex.ru/",
			},
			request: "/shortenLink1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			getMethod(w, request)
			result := w.Result()
			fmt.Println(result.Header, result.StatusCode)
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
