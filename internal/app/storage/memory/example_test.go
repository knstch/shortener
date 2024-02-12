package storage

import (
	"context"
	"fmt"
)

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