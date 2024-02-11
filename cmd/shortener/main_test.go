package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knstch/shortener/cmd/config"
	"github.com/knstch/shortener/internal/app/handler"
	"github.com/knstch/shortener/internal/app/logger"
	"github.com/knstch/shortener/internal/app/router"
	dbconnect "github.com/knstch/shortener/internal/app/storage/DBConnect"
	memory "github.com/knstch/shortener/internal/app/storage/memory"
	"github.com/knstch/shortener/internal/app/storage/psql"
	"github.com/stretchr/testify/assert"
)

func linkGenerator(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	stringRes := string(result)
	return string(stringRes + "_long_link")
}

type user struct {
	login      string
	shortLinks []string
	longLinks  []string
	cookie     *http.Cookie
}

var userOne user

type want struct {
	statusCode  int
	contentType string
	body        string
}

type request struct {
	contentType string
	body        string
}

func TestPostLink(t *testing.T) {
	config.ParseConfig()
	var storage handler.Storage
	var ping handler.PingChecker
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
		ping = dbconnect.NewDBConnection(db)
	} else {
		storage = memory.NewMemStorage()
	}
	h := handler.NewHandler(storage, ping)

	router := router.RequestsRouter(h)

	longLink := linkGenerator(10)
	userOne.longLinks = append(userOne.longLinks, longLink)

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 regular post",
			want: want{
				statusCode:  201,
				contentType: "text/plain",
			},
			reqest: request{
				contentType: "text/plain",
				body:        longLink,
			},
		},
		{
			name: "#2 repeat post",
			want: want{
				statusCode:  409,
				contentType: "text/plain",
			},
			reqest: request{
				contentType: "text/plain",
				body:        longLink,
			},
		},
		// {
		// 	name: "#3 post normal link",
		// 	want: want{
		// 		statusCode:  200,
		// 		contentType: "text/plain",
		// 	},
		// 	reqest: request{
		// 		contentType: "text/plain",
		// 		body:        "https://www.google.com/",
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			if userOne.cookie == nil {
				getCookieRes := httptest.NewRecorder()
				getCookie := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", bytes.NewBuffer([]byte(linkGenerator(10))))
				getCookie.Header.Set("Content-Type", tt.reqest.contentType)
				router.ServeHTTP(getCookieRes, getCookie)
				cookies := getCookieRes.Result().Cookies()
				for _, cookie := range cookies {
					userOne.cookie = cookie
				}
			}
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.statusCode, rr.Code)
			userOne.shortLinks = append(userOne.shortLinks, rr.Body.String())
		})
	}
}

func BenchmarkPostLink(b *testing.B) {
	var storage handler.Storage
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
	} else {
		storage = memory.NewMemStorage()
	}
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		storage.PostLink(ctx, linkGenerator(5), config.ReadyConfig.BaseURL, 0)
	}
}

func TestGetLink(t *testing.T) {
	var storage handler.Storage
	var ping handler.PingChecker
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
		ping = dbconnect.NewDBConnection(db)
	} else {
		storage = memory.NewMemStorage()
	}
	h := handler.NewHandler(storage, ping)

	router := router.RequestsRouter(h)

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 regular get",
			want: want{
				statusCode:  404,
				contentType: "text/plain; charset=utf-8",
			},
			reqest: request{
				body: userOne.shortLinks[0],
			},
		},
		{
			name: "#2 empty get",
			want: want{
				statusCode:  405,
				contentType: "",
			},
			reqest: request{
				body: "",
			},
		},
		// {
		// 	name: "#3 normal link get",
		// 	want: want{
		// 		statusCode:  200,
		// 		contentType: "text/html",
		// 	},
		// 	reqest: request{
		// 		body: userOne.shortLinks[1],
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/"+tt.reqest.body, bytes.NewBuffer([]byte("")))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.statusCode, rr.Code)
			userOne.shortLinks = append(userOne.shortLinks, rr.Body.String())
		})
	}
}

func BenchmarkFindLink(b *testing.B) {
	var storage handler.Storage
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
	} else {
		storage = memory.NewMemStorage()
	}
	for i := 0; i < b.N; i++ {
		storage.FindLink(linkGenerator(5))
	}
}

func TestPostLinkJSON(t *testing.T) {
	var storage handler.Storage
	var ping handler.PingChecker
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
		ping = dbconnect.NewDBConnection(db)
	} else {
		storage = memory.NewMemStorage()
	}
	h := handler.NewHandler(storage, ping)

	router := router.RequestsRouter(h)

	longLink := linkGenerator(10)
	userOne.longLinks = append(userOne.longLinks, longLink)

	var shortLink handler.Result

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 regular post JSON",
			want: want{
				statusCode:  201,
				contentType: "application/json",
			},
			reqest: request{
				contentType: "application/json",
				body:        `{"url": "` + longLink + `"}`,
			},
		},
		{
			name: "#2 repeat post JSON",
			want: want{
				statusCode:  409,
				contentType: "application/json",
			},
			reqest: request{
				contentType: "application/json",
				body:        `{"url": "` + longLink + `"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			req.AddCookie(userOne.cookie)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.statusCode, rr.Code)

			body, err := io.ReadAll(rr.Body)
			if err != nil {
				logger.ErrorLogger("Error during opening body: ", err)
			}

			json.Unmarshal(body, &shortLink)
			userOne.shortLinks = append(userOne.shortLinks, rr.Body.String())
		})
	}
}

func TestPostLinkJSONBatch(t *testing.T) {
	var storage handler.Storage
	var ping handler.PingChecker
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
		ping = dbconnect.NewDBConnection(db)
	} else {
		storage = memory.NewMemStorage()
	}
	h := handler.NewHandler(storage, ping)

	router := router.RequestsRouter(h)

	longLinkOne := linkGenerator(10)
	longLinkTwo := linkGenerator(10)
	userOne.longLinks = append(userOne.longLinks, longLinkOne, longLinkTwo)

	var shortLinkBatch []handler.ShortLink

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 regular post JSON batch",
			want: want{
				statusCode:  201,
				contentType: "application/json",
			},
			reqest: request{
				contentType: "application/json",
				body:        `[{"original_url": "` + longLinkOne + `","correlation_id": "` + longLinkOne + "id" + `"},{"original_url": "` + longLinkTwo + `","correlation_id": "` + longLinkTwo + "id" + `"}]`,
			},
		},
		{
			name: "#2 repeat post JSON batch",
			want: want{
				statusCode:  409,
				contentType: "application/json",
			},
			reqest: request{
				contentType: "application/json",
				body:        `[{"original_url": "` + longLinkOne + `","correlation_id": "` + longLinkOne + "id" + `"},{"original_url": "` + longLinkTwo + `","correlation_id": "` + longLinkTwo + "id" + `"}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/batch", bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			req.AddCookie(userOne.cookie)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.statusCode, rr.Code)

			body, err := io.ReadAll(rr.Body)
			if err != nil {
				logger.ErrorLogger("Error during opening body: ", err)
			}

			json.Unmarshal(body, &shortLinkBatch)
			for _, link := range shortLinkBatch {
				userOne.shortLinks = append(userOne.shortLinks, link.Result)
			}
			fmt.Println("checking links", len(userOne.shortLinks))
		})
	}
}

func BenchmarkGetURLsByID(b *testing.B) {
	var storage handler.Storage
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
	} else {
		storage = memory.NewMemStorage()
	}
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		storage.GetURLsByID(ctx, 0, config.ReadyConfig.BaseURL)
	}
}

func BenchmarkPing(b *testing.B) {
	var connect dbconnect.DBConnection
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
		}
		connect = *dbconnect.NewDBConnection(db)
	}
	for i := 0; i < b.N; i++ {
		connect.Ping()
	}
}
