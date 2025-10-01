package main

import (
	"log"

	"github.com/MarkelovSergey/url-shorter/internal/app"
	"github.com/MarkelovSergey/url-shorter/internal/config"
)

func main() {
	cfg := config.ParseFlags()
	app := app.New(cfg)

	if err := app.Run(); err != nil {
		log.Fatal("Application error:", err)
	}
}
