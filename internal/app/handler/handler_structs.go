package handler

import (
	"context"

	"github.com/knstch/shortener/internal/app/common"
)

// IStorage объединяет методы для взаимодействия с БД.
type Storager interface {
	FindLink(ctx context.Context, url string) (string, bool, error)
	PostLink(ctx context.Context, longLink string, URLaddr string, UserID int) (string, error)
	GetURLsByID(ctx context.Context, id int, URLaddr string) ([]common.URLs, error)
	DeleteURLs(ctx context.Context, id int, shortURLs []string) error
	GetStats(ctx context.Context) ([]byte, error)
}

// PingChecker имеет 1 метод для проверки соединения с БД.
type PingChecker interface {
	Ping() error
}

// Handler объединяет интерфейсы для проверки соединения и взаимодействия с БД
type Handler struct {
	s Storager
	p PingChecker
}

type link struct {
	URL string `json:"url"`
}

// Result используется для декодирования данных из JSON формата.
type Result struct {
	Result string `json:"result"`
}

type originalLink struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

// ShortLink используется для декодирования данных из JSON формата.
type ShortLink struct {
	Result        string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}
