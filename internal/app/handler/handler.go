package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/knstch/shortener/cmd/config"
	"github.com/knstch/shortener/internal/app/logger"
)

type Storage interface {
	FindLink(url string) (string, error)
	PostLink(ctx context.Context, reqBody string, URLaddr string) (string, error)
}

type PingChecker interface {
	Ping() error
}

type Handler struct {
	s Storage
	p PingChecker
}

func NewHandler(s Storage, p PingChecker) *Handler {
	return &Handler{s: s, p: p}
}

// Вызывается при использовании метода POST, передает данные
// в функцию postURL для записи данных в хранилище и пишет
// ответ сервера, когда все записано
func (h *Handler) PostURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during reading body: ", err)
		res.WriteHeader(500)
		return
	}
	returnedShortLink, err := h.s.PostLink(req.Context(), string(body), config.ReadyConfig.BaseURL)
	if err != nil {
		logger.ErrorLogger("Error during reading body: ", err)
		res.WriteHeader(500)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(200)
	res.Write([]byte(returnedShortLink))
}

// Структура для приема URL
type link struct {
	URL string `json:"url"`
}

// Структура для записи в json
type result struct {
	Result string `json:"result"`
}

// Передаем json-объект и получаем в ответе короткий URL в виде json-объекта
func (h *Handler) PostURLJSON(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}
	var longLink link
	json.Unmarshal(body, &longLink)
	shortenURL, err := h.s.PostLink(req.Context(), longLink.URL, config.ReadyConfig.BaseURL)
	if err != nil {
		// todo
	}
	var resultJSON = result{
		Result: shortenURL,
	}
	resp, err := json.Marshal(resultJSON)
	if err != nil {
		logger.ErrorLogger("Fail during convertion to json: ", err)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	res.Write(resp)
}

type originalLink struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type shortLink struct {
	Result        string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}

func (h *Handler) PostBatch(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(201)
	var originalRequest []originalLink
	var shortenResponse []shortLink

	err := json.NewDecoder(req.Body).Decode(&originalRequest)
	if err != nil {
		logger.ErrorLogger("Failed to read json: ", err)
	}

	for i := range originalRequest {
		returnedShortLink, err := h.s.PostLink(req.Context(), originalRequest[i].OriginalURL, config.ReadyConfig.BaseURL)
		if err != nil {
			// todo
		}
		shortenResponse = append(shortenResponse,
			shortLink{
				Result:        returnedShortLink,
				CorrelationID: originalRequest[i].CorrelationID,
			})
	}

	response, err := json.Marshal(shortenResponse)
	if err != nil {
		logger.ErrorLogger("Failed to marshal json: ", err)
	}

	res.Write(response)
}

// Вызываем для передачи данных в функцию getURL
// и написания ответа в зависимости от ответа getURL
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	url := chi.URLParam(req, "url")
	shortenURL, err := h.s.FindLink(url)
	if err != nil {
		logger.ErrorLogger("Can't find link: ", err)
		http.Error(res, "Bad Request", 500)
		return
	}

	if shortenURL == "" {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Location", shortenURL)
	res.WriteHeader(307)
	res.Write([]byte(shortenURL))
}

// Проверяем соединение с базой данных
func (h *Handler) PingDB(res http.ResponseWriter, req *http.Request) {
	if h.p != nil {
		if err := h.p.Ping(); err != nil {
			http.Error(res, "Can't connect to DB", http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Connection is set"))
}
