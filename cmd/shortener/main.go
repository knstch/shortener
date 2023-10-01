package main

import (
	"io"
	"net/http"
	"strconv"
)

var storage = make(map[string]string)
var counter int

func getMethod(res http.ResponseWriter, req *http.Request) {
	for k, v := range storage {
		if "/"+v == req.URL.String() {
			res.Header().Set("Content-Type", "text/plain")
			res.Header().Add("Location", k)
			res.WriteHeader(307)
			res.Write([]byte(k))
			return
		}
	}
	http.Error(res, "Bad Request", http.StatusBadRequest)
}

func postMethod(res http.ResponseWriter, req *http.Request) {
	counter++
	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	storage[string(body)] = "shortenLink" + strconv.Itoa(counter)
	shortenLink := "http://localhost:8080/" + storage[string(body)]
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(201)
	res.Write([]byte(shortenLink))
}

func mainPage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		getMethod(res, req)
	case http.MethodPost:
		postMethod(res, req)
	default:
		http.Error(res, "Bad Request", http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainPage)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
