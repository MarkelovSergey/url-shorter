package config

import (
	"flag"
	"os"
)

const (
	serverAddressEnv   = "SERVER_ADDRESS"
	baseURLEnv         = "BASE_URL"
	fileStoragePathEnv = "FILE_STORAGE_PATH"
	databaseDSNEnv     = "DATABASE_DSN"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func New(serverAddr, baseURL, fileStoragePath, databaseDSN string) Config {
	return Config{
		ServerAddress:   serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
	}
}

func ParseFlags() Config {
	serverAddr := flag.String("a", ":8080", "HTTP server address (e.g. localhost:8888)")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	fileStoragePath := flag.String("f", "/var/lib/url-shorter/short-url-db.json", "file storage path")
	databaseDSN := flag.String("d", "", "database connection string")
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

	finalDatabaseDSN := *databaseDSN
	if envDatabaseDSN, ok := os.LookupEnv(databaseDSNEnv); ok {
		finalDatabaseDSN = envDatabaseDSN
	}

	return New(finalServerAddr, finalBaseURL, finalFileStoragePath, finalDatabaseDSN)
}
