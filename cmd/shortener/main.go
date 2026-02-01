package main

import (
	"fmt"
	"log"

	"github.com/MarkelovSergey/url-shorter/internal/app"
	"github.com/MarkelovSergey/url-shorter/internal/config"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	cfg := config.ParseFlags()
	app := app.New(cfg)

	if err := app.Run(); err != nil {
		log.Fatal("Application error:", err)
	}
}

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
