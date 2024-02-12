package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func PostLink(t *testing.T) {
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
	storage := NewMemStorage()
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testShortLink, _ := storage.PostLink(context.Background(), tests[i].request, "http://localhost:8080", 0)
			assert.Equal(t, testShortLink, tests[i].want)
		})
	}
}

func TestFindLink(t *testing.T) {

	storage := NewMemStorage()
	storage.Data["shortenLink1"] = "https://practicum.yandex.ru/"
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
			longLink, _, _ := storage.FindLink(tests[i].request)
			assert.Equal(t, longLink, tests[i].want)
		})
	}
}

var testMemStorage = NewMemStorage()

func ExampleMemStorage_PostLink() {
	ctx := context.Background()

	longLinkToSave := "https://practicum.yandex.ru/"

	shortLink, _ := testMemStorage.PostLink(ctx, longLinkToSave, "http://localhost:8080", 0)
	fmt.Println(shortLink)

	// Output:
	// http://localhost:8080/shortenLink1
}

func ExampleMemStorage_FindLink() {

	longLink, _, _ := testMemStorage.FindLink("shortenLink1")
	fmt.Println(longLink)

	// Output:
	// https://practicum.yandex.ru/
}