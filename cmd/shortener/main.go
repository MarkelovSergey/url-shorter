package main

import (
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/handler"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"

	"github.com/go-chi/chi/v5"
)

func main() {
	urlShorterRepo := urlshorterrepository.New()
	urlShorterService := urlshorterservice.New(urlShorterRepo)

	handler := handler.New(urlShorterService)
	r := chi.NewRouter()

	r.Post("/", handler.CreateHandler)
	r.Get("/{id}", handler.ReadHandler)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
