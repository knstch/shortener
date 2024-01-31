package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	config "github.com/knstch/shortener/cmd/config"
	logger "github.com/knstch/shortener/internal/app/logger"
)

var pgErr *pgconn.PgError

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
	}
	returnedShortLink, err := h.s.PostLink(req.Context(), string(body), config.ReadyConfig.BaseURL)
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(409)
		res.Write([]byte(returnedShortLink))
		return
	} else if err != nil {
		logger.ErrorLogger("Error posing link: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(201)
	res.Write([]byte(returnedShortLink))
}

// Функция принимает ссылку в json и отдает короткую в json
func (h *Handler) PostLongLinkJSON(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}
	var longLink link
	json.Unmarshal(body, &longLink)

	shortenURL, err := h.s.PostLink(req.Context(), longLink.URL, config.ReadyConfig.BaseURL)
	var resultJSON = result{
		Result: shortenURL,
	}
	resp, _ := json.Marshal(resultJSON)
	fmt.Printf("Shorten duplicate: %v\n", shortenURL)
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(409)
		res.Write([]byte(resp))
		return
	} else if err != nil {
		logger.ErrorLogger("Error posting link: %v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(201)
	res.Write(resp)
}

func (h *Handler) PostBatch(res http.ResponseWriter, req *http.Request) {
	var originalRequest []originalLink
	var shortenResponse []shortLink

	err := json.NewDecoder(req.Body).Decode(&originalRequest)
	if err != nil {
		logger.ErrorLogger("Failed to read json: ", err)
	}

	for i := range originalRequest {
		returnedShortLink, err := h.s.PostLink(req.Context(), originalRequest[i].OriginalURL, config.ReadyConfig.BaseURL)
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(409)
			res.Write([]byte(returnedShortLink))
			return
		} else if err != nil {
			logger.ErrorLogger("Error posing link: ", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
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
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(201)
	res.Write(response)
}

// Вызываем для передачи данных в функцию getURL
// и написания ответа в зависимости от ответа getURL
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	url = strings.Trim(url, "/")
	fmt.Printf("URL: %v\n", url)
	shortenURL, err := h.s.FindLink(url)
	if err != nil {
		logger.ErrorLogger("Can't find link: ", err)
		http.Error(res, "Bad Request", http.StatusInternalServerError)
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
