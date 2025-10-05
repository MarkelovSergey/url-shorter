package config

import (
	"flag"
	"os"
)

const (
	serverAddressEnv = "SERVER_ADDRESS"
	baseURLEnv       = "BASE_URL"
)

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

	finalServerAddr := *serverAddr
	if envServerAddr := os.Getenv(serverAddressEnv); envServerAddr != "" {
		finalServerAddr = envServerAddr
	}

	finalBaseURL := *baseURL
	if envBaseURL := os.Getenv(baseURLEnv); envBaseURL != "" {
		finalBaseURL = envBaseURL
	}

	return *New(finalServerAddr, finalBaseURL)
}
