package main

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	getMethod "github.com/knstch/shortener_url/internal/app/getMethod"
	postMethod "github.com/knstch/shortener_url/internal/app/postMethod"
)

func mainPage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		url := chi.URLParam(req, "url")
		if shortenURL := getMethod.GetMethod(url, postMethod.StorageURLs); shortenURL != "" {
			res.Header().Set("Content-Type", "text/plain")
			res.Header().Set("Location", shortenURL)
			res.WriteHeader(307)
			res.Write([]byte(shortenURL))
		} else {
			http.Error(res, "Bad Request", http.StatusBadRequest)
		}

	case http.MethodPost:
		body, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(201)
		res.Write([]byte(postMethod.PostMethod(string(body), &postMethod.StorageURLs)))
	default:
		http.Error(res, "Bad Request", http.StatusBadRequest)
	}
}

func main() {
	r := chi.NewRouter()
	r.Get("/{url}", mainPage)
	r.Post("/", mainPage)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
