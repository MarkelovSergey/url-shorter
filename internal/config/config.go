package config

import "flag"

type Config struct {
	ServerAddress string
	BaseURL       string
}

func New(serverAddr, baseURL string) *Config {
	return &Config{
		ServerAddress: serverAddr,
		BaseURL:       baseURL,
	}
}

func ParseFlags() Config {
	serverAddr := flag.String("a", ":8080", "HTTP server address (e.g. localhost:8888)")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	flag.Parse()

	return *New(*serverAddr, *baseURL)
}
