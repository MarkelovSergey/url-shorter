package main

import (
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/handler"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
)

func main() {
	urlShorterRepo := urlshorterrepository.New()
	urlShorterService := urlshorterservice.New(urlShorterRepo)

	handler := handler.New(*urlShorterService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.CreateHandler)
	mux.HandleFunc("/{id}", handler.ReadHandler)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
