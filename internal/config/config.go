// Package config содержит конфигурацию приложения.
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
	auditFileEnv       = "AUDIT_FILE"
	auditURLEnv        = "AUDIT_URL"
)

// Config содержит настройки приложения.
type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	AuditFile       string
	AuditURL        string
}

// New создает новый экземпляр конфигурации с заданными параметрами.
func New(serverAddr, baseURL, fileStoragePath, databaseDSN, auditFile, auditURL string) Config {
	return Config{
		ServerAddress:   serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
		AuditFile:       auditFile,
		AuditURL:        auditURL,
	}
}

// ParseFlags парсит флаги командной строки и переменные окружения.
func ParseFlags() Config {
	serverAddr := flag.String("a", ":8080", "HTTP server address (e.g. localhost:8888)")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	fileStoragePath := flag.String("f", "/var/lib/url-shorter/short-url-db.json", "file storage path")
	databaseDSN := flag.String("d", "", "database connection string")
	auditFile := flag.String("audit-file", "", "path to audit log file")
	auditURL := flag.String("audit-url", "", "URL of remote audit server")
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

	finalAuditFile := *auditFile
	if envAuditFile, ok := os.LookupEnv(auditFileEnv); ok {
		finalAuditFile = envAuditFile
	}

	finalAuditURL := *auditURL
	if envAuditURL, ok := os.LookupEnv(auditURLEnv); ok {
		finalAuditURL = envAuditURL
	}

	return New(finalServerAddr, finalBaseURL, finalFileStoragePath, finalDatabaseDSN, finalAuditFile, finalAuditURL)
}
