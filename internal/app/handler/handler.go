// Модуль handler отвечает за хедлеры.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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

// NewHandler возвращает интерфейс Handler.
func NewHandler(s Storager, p PingChecker) *Handler {
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
		_, err = res.Write([]byte(returnedShortLink))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
		return
	} else if err != nil {
		logger.ErrorLogger("Error posing link: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(201)
	_, err = res.Write([]byte(returnedShortLink))
	if err != nil {
		logger.ErrorLogger("Can't write response: ", err)
	}
}

// PostLongLinkJSON выполняет ту же функцию, что и PostURL, но принимает
// и отдает данные в JSON формате.
func (h *Handler) PostLongLinkJSON(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}
	var longLink link
	err = json.Unmarshal(body, &longLink)
	if err != nil {
		logger.ErrorLogger("Error unmarshallig JSON: ", err)
	}

	UserID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Error getting cookie: ", err)
		cookieError := Result{
			Result: "Error operating with cookie",
		}
		resp, _ := json.Marshal(cookieError)
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusInternalServerError)
		_, err = res.Write([]byte(resp))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
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
		_, err = res.Write([]byte(resp))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
		return
	} else if err != nil {
		logger.ErrorLogger("Error posting link: %v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(201)
	_, err = res.Write(resp)
	if err != nil {
		logger.ErrorLogger("Can't write response: ", err)
	}
}

// PostBatch выполняет ту же функцию, что и PostURL, но может
// принимать множество адресов.
func (h *Handler) PostBatch(res http.ResponseWriter, req *http.Request) {
	var originalRequest []originalLink
	var shortenResponse []ShortLink

	statusCode := http.StatusCreated

	err := json.NewDecoder(req.Body).Decode(&originalRequest)
	if err != nil {
		logger.ErrorLogger("Failed to read json: ", err)
	}

	UserID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Error getting cookie: ", err)
	}

	for i := range originalRequest {
		var returnedShortLink string
		returnedShortLink, err = h.s.PostLink(req.Context(), originalRequest[i].OriginalURL, config.ReadyConfig.BaseURL, UserID)
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
	_, err = res.Write(response)
	if err != nil {
		logger.ErrorLogger("Can't write response: ", err)
	}
}

// GetURL принимает короткий адрес в path и если был найден
// длинный адрес, то перенаправляет клиента по этой ссылке.
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	url = strings.Trim(url, "/")

	shortenURL, deleteStatus, err := h.s.FindLink(req.Context(), url)
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
		_, err = res.Write([]byte("Deleted URL"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Location", shortenURL)
	res.WriteHeader(307)
	_, err = res.Write([]byte(shortenURL))
	if err != nil {
		logger.ErrorLogger("Can't write response: ", err)
	}
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
	_, err := res.Write([]byte("Connection is set"))
	if err != nil {
		logger.ErrorLogger("Can't write response: ", err)
	}
}

// GetUserLinks отдает клиенту все ссылки загруженные им проверяя
// ID с помощью куки.
func (h *Handler) GetUserLinks(res http.ResponseWriter, req *http.Request) {

	userID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Unauthorized access : ", err)
		res.WriteHeader(401)
		_, err = res.Write([]byte("Unauthorized"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
		return
	}
	userURLs, err := h.s.GetURLsByID(req.Context(), userID, config.ReadyConfig.BaseURL)
	if err != nil {
		logger.ErrorLogger("Error getting URLs by ID", err)
	}

	res.Header().Set("Content-Type", "application/json")
	if string(userURLs) == "null" {
		res.WriteHeader(200)
		_, err = res.Write([]byte("No content"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
	} else {
		res.WriteHeader(200)
		_, err = res.Write(userURLs)
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
	}
}

// DeleteLinks принимает множество ссылок от пользователя и удаляет их,
// если они были загруженным клиентом отправляющим этот запрос.
func (h *Handler) DeleteLinks(res http.ResponseWriter, req *http.Request) {
	userID, err := cookies.CheckCookieForID(res, req)
	if err != nil {
		logger.ErrorLogger("Error getting cookie: ", err)
		res.WriteHeader(401)
		_, err = res.Write([]byte("You have no links to delete"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
		return
	}

	var URLs []string

	err = json.NewDecoder(req.Body).Decode(&URLs)
	if err != nil {
		logger.ErrorLogger("Failed to read json: ", err)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusInternalServerError)
		_, err = res.Write([]byte("Failed to read JSON"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
	}

	err = h.s.DeleteURLs(req.Context(), userID, URLs)
	if err != nil {
		logger.ErrorLogger("Error deleting links", err)
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(202)
	_, err = res.Write([]byte("Deleted"))
	if err != nil {
		logger.ErrorLogger("Can't write response: ", err)
	}
}

// GetStatsHandler выдает статистику по пользователям и ссылкам.
// Информация доступна только доверенным пользователям.
func (h *Handler) GetStatsHandler(res http.ResponseWriter, req *http.Request) {
	clientIP, err := subnetChecker(req)
	if err != nil {
		logger.ErrorLogger("Failed to read clients IP: ", err)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusInternalServerError)
		_, err = res.Write([]byte("Failed to read clients IP"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
	}
	if clientIP != config.ReadyConfig.TrustedSubnet {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusForbidden)
		_, err = res.Write([]byte("Доступ запрещен"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
		return
	}
	stats, err := h.s.GetStats(req.Context())
	if err != nil {
		logger.ErrorLogger("Failed to load stats: ", err)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusInternalServerError)
		_, err = res.Write([]byte("Failed to load stats"))
		if err != nil {
			logger.ErrorLogger("Can't write response: ", err)
		}
	}
	res.Header().Set("Content-Type", "tapplication/json")
	res.WriteHeader(200)
	_, err = res.Write(stats)
	if err != nil {
		logger.ErrorLogger("Can't write response: ", err)
	}
}

func subnetChecker(req *http.Request) (string, error) {
	ipStr := req.Header.Get("X-Real-IP")
	// парсим ip
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// если заголовок X-Real-IP пуст, пробуем X-Forwarded-For
		// этот заголовок содержит адреса отправителя и промежуточных прокси
		// в виде 203.0.113.195, 70.41.3.18, 150.172.238.178
		ips := req.Header.Get("X-Forwarded-For")
		// разделяем цепочку адресов
		ipStrs := strings.Split(ips, ",")
		// интересует только первый
		ipStr = ipStrs[0]
		// парсим
		ip = net.ParseIP(ipStr)
		if ip == nil {
			ipStr, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				return "", err
			}
			ip = net.ParseIP(ipStr)
		}
	}
	if ip == nil {
		return "", fmt.Errorf("невозможно определить ip")
	}
	return ip.String(), nil
}
