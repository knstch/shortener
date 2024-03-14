package common

import (
	"math/rand"
)

// Stats нужен для конвертации статистики в JSON.
type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// URLs используется для кодирования данных в JSON формат.
type URLs struct {
	LongLink  string `json:"original_url"`
	ShortLink string `json:"short_url"`
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
