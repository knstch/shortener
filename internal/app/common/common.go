package common

import (
	"math/rand"
)

type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// Генератор короткой ссылки
func ShortLinkGenerator(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
