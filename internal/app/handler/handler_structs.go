package handler

import "context"

// IStorage объединяет методы для взаимодействия с БД.
type IStorage interface {
	FindLink(url string) (string, bool, error)
	PostLink(ctx context.Context, longLink string, URLaddr string, UserID int) (string, error)
	GetURLsByID(ctx context.Context, id int, URLaddr string) ([]byte, error)
	DeleteURLs(ctx context.Context, id int, shortURLs []string) error
}

// PingChecker имеет 1 метод для проверки соединения с БД.
type PingChecker interface {
	Ping() error
}

// Handler объединяет интерфейсы для проверки соединения и взаимодействия с БД
type Handler struct {
	s IStorage
	p PingChecker
}

type link struct {
	URL string `json:"url"`
}

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
