package main

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	getMethod "github.com/knstch/shortener_url/internal/app/getMethod"
	postMethod "github.com/knstch/shortener_url/internal/app/postMethod"
)

// Вызываем для передачи данных в функцию GetMethod
// и написания ответа в зависимости от ответа GetMethod
func getURL(res http.ResponseWriter, req *http.Request) {
	url := chi.URLParam(req, "url")
	if shortenURL := getMethod.GetMethod(url, postMethod.StorageURLs); shortenURL != "" {
		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Location", shortenURL)
		res.WriteHeader(307)
		res.Write([]byte(shortenURL))
	} else {
		http.Error(res, "Bad Request", http.StatusBadRequest)
	}
}

// Вызывается при использовании метода POST, передает данные
// в функцию PostMethod для записи данных в хранилище и пишет
// ответ сервера, когда все записано 
func postURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(201)
	res.Write([]byte(postMethod.PostMethod(string(body), &postMethod.StorageURLs)))
}

// Роутер запросов
func RequestsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{url}", getURL)
	r.Post("/", postURL)
	return r
}

func main() {
	err := http.ListenAndServe(":8080", RequestsRouter())
	if err != nil {
		panic(err)
	}
}
