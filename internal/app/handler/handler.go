// Модуль handler отвечает за хедлеры.
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
	cookies "github.com/knstch/shortener/internal/app/cookies"
	logger "github.com/knstch/shortener/internal/app/logger"
	memStorage "github.com/knstch/shortener/internal/app/storage/memory"
)

var pgErr *pgconn.PgError
var memStorageIntegrityErr *memStorage.IntegrityError

func NewHandler(s IStorage, p PingChecker) *Handler {
	return &Handler{s: s, p: p}
}

// PostURL используется для записи нового URL и предоставления короткой ссылки,
// если URL уже записан, то возвращается код 409 и дается коротка ссылка.
func (h *Handler) PostURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during reading body: ", err)
	}

	UserID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Error getting cookie: ", err)
	}

	returnedShortLink, err := h.s.PostLink(req.Context(), string(body), config.ReadyConfig.BaseURL, UserID)
	if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)) || errors.As(err, &memStorageIntegrityErr) {
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

// PostLongLinkJSON выполняет ту же функцию, что и PostURL, но принимает
// и отдает данные в JSON формате.
func (h *Handler) PostLongLinkJSON(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}
	var longLink link
	json.Unmarshal(body, &longLink)

	UserID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Error getting cookie: ", err)
	}

	shortenURL, err := h.s.PostLink(req.Context(), longLink.URL, config.ReadyConfig.BaseURL, UserID)
	var resultJSON = Result{
		Result: shortenURL,
	}
	resp, _ := json.Marshal(resultJSON)
	fmt.Printf("Shorten duplicate: %v\n", shortenURL)
	if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)) || errors.As(err, &memStorageIntegrityErr) {
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

// PostBatch выполняет ту же функцию, что и PostURL, но может 
// принимать множество адресов.
func (h *Handler) PostBatch(res http.ResponseWriter, req *http.Request) {
	var originalRequest []originalLink
	var shortenResponse []ShortLink

	statusCode := 201

	err := json.NewDecoder(req.Body).Decode(&originalRequest)
	if err != nil {
		logger.ErrorLogger("Failed to read json: ", err)
	}

	UserID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Error getting cookie: ", err)
	}

	for i := range originalRequest {

		returnedShortLink, err := h.s.PostLink(req.Context(), originalRequest[i].OriginalURL, config.ReadyConfig.BaseURL, UserID)
		if (errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)) || errors.As(err, &memStorageIntegrityErr) {
			statusCode = 409
		} else if err != nil {
			logger.ErrorLogger("Error posing link: ", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		shortenResponse = append(shortenResponse,
			ShortLink{
				Result:        returnedShortLink,
				CorrelationID: originalRequest[i].CorrelationID,
			})
	}

	response, err := json.Marshal(shortenResponse)
	if err != nil {
		logger.ErrorLogger("Failed to marshal json: ", err)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	res.Write(response)
}

// GetURL принимает короткий адрес в path и если был найден
// длинный адрес, то перенаправляет клиента по этой ссылке.
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	url = strings.Trim(url, "/")

	shortenURL, deleteStatus, err := h.s.FindLink(url)
	if err != nil {
		logger.ErrorLogger("Can't find link: ", err)
		http.Error(res, "Bad Request", http.StatusInternalServerError)
		return
	}

	if shortenURL == "" {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}

	if deleteStatus {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(410)
		res.Write([]byte("Deleted URL"))
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Location", shortenURL)
	res.WriteHeader(307)
	res.Write([]byte(shortenURL))
}

// PingDB используется для проверки соединения с БД.
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

// GetUserLinks отдает клиенту все ссылки загруженные им проверяя
// ID с помощью куки.
func (h *Handler) GetUserLinks(res http.ResponseWriter, req *http.Request) {

	userID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Unauthorized access : ", err)
		res.WriteHeader(401)
		res.Write([]byte("Unauthorized"))
		return
	}
	userURLs, err := h.s.GetURLsByID(req.Context(), userID, config.ReadyConfig.BaseURL)
	if err != nil {
		logger.ErrorLogger("Error getting URLs by ID", err)
	}

	res.Header().Set("Content-Type", "application/json")
	if string(userURLs) == "null" {
		res.WriteHeader(200)
		res.Write([]byte("No content"))
	} else {
		res.WriteHeader(200)
		res.Write(userURLs)
	}
}

// DeleteLinks принимает множество ссылок от пользователя и удаляет их,
// если они были загруженным клиентом отправляющим этот запрос.
func (h *Handler) DeleteLinks(res http.ResponseWriter, req *http.Request) {
	userID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Error getting cookie: ", err)
		res.WriteHeader(401)
		res.Write([]byte("You have no links to delete"))
		return
	}

	var URLs []string

	err = json.NewDecoder(req.Body).Decode(&URLs)
	if err != nil {
		logger.ErrorLogger("Failed to read json: ", err)
	}

	err = h.s.DeleteURLs(req.Context(), userID, URLs)
	if err != nil {
		logger.ErrorLogger("Error deleting links", err)
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(202)
	res.Write([]byte("Deleted"))
}
