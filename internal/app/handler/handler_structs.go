package handler

import "context"

type Storage interface {
	FindLink(url string) (string, bool, error)
	PostLink(ctx context.Context, longLink string, URLaddr string, UserID int) (string, error)
	GetURLsByID(ctx context.Context, id int, URLaddr string) ([]byte, error)
	DeleteURLs(ctx context.Context, id int, shortURLs []string) error
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
type Result struct {
	Result string `json:"result"`
}

type originalLink struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type ShortLink struct {
	Result        string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}
