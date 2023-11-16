package handler

import "context"

type Storage interface {
	FindLink(url string) (string, error)
	PostLink(ctx context.Context, longLink string, URLaddr string) (string, error)
}

type PingChecker interface {
	Ping() error
}

type Handler struct {
	s Storage
	p PingChecker
}

// Структура для приема URL
type link struct {
	URL string `json:"url"`
}

// Структура для записи в json
type result struct {
	Result string `json:"result"`
}

type originalLink struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type shortLink struct {
	Result        string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}
