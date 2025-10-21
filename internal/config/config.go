package config

import (
	"flag"
	"os"
)

const (
	serverAddressEnv   = "SERVER_ADDRESS"
	baseURLEnv         = "BASE_URL"
	fileStoragePathEnv = "FILE_STORAGE_PATH"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func New(serverAddr, baseURL, fileStoragePath string) Config {
	return Config{
		ServerAddress:   serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
	}
}

func ParseFlags() Config {
	serverAddr := flag.String("a", ":8080", "HTTP server address (e.g. localhost:8888)")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	fileStoragePath := flag.String("f", "/var/lib/url-shorter/short-url-db.json", "file storage path")
	flag.Parse()

	finalServerAddr := *serverAddr
	if envServerAddr, ok := os.LookupEnv(serverAddressEnv); ok {
		finalServerAddr = envServerAddr
	}

	finalBaseURL := *baseURL
	if envBaseURL, ok := os.LookupEnv(baseURLEnv); ok {
		finalBaseURL = envBaseURL
	}

	finalFileStoragePath := *fileStoragePath
	if envFileStoragePath, ok := os.LookupEnv(fileStoragePathEnv); ok {
		finalFileStoragePath = envFileStoragePath
	}

	return New(finalServerAddr, finalBaseURL, finalFileStoragePath)
}
