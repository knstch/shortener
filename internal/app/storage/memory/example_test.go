package storage

import (
	"context"
	"fmt"
)

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
	ctx := context.Background()
	longLink, _, _ := testMemStorage.FindLink(ctx, "shortenLink1")
	fmt.Println(longLink)

	// Output:
	// https://practicum.yandex.ru/
}
