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

// ServerConfig содержит настройки HTTP-сервера.
type ServerConfig struct {
	// Address - адрес HTTP-сервера (например, ":8080" или "localhost:8888")
	Address string
	// BaseURL - базовый URL для создания коротких ссылок (например, "http://localhost:8080")
	BaseURL string
}

// StorageConfig содержит настройки хранилища данных.
type StorageConfig struct {
	// FilePath - путь к файлу для хранения URL (если не используется PostgreSQL)
	FilePath string
}

// DatabaseConfig содержит настройки подключения к базе данных.
type DatabaseConfig struct {
	// DSN - строка подключения к PostgreSQL (если задана, используется вместо файлового хранилища)
	DSN string
}

// AuditConfig содержит настройки системы аудита.
type AuditConfig struct {
	// FilePath - путь к файлу для записи событий аудита
	FilePath string
	// URL - URL удаленного сервера для отправки событий аудита
	URL string
}

// Config содержит настройки приложения.
type Config struct {
	// Server - настройки HTTP-сервера
	Server ServerConfig
	// Storage - настройки файлового хранилища
	Storage StorageConfig
	// Database - настройки подключения к базе данных
	Database DatabaseConfig
	// Audit - настройки системы аудита
	Audit AuditConfig
}

// New создает новый экземпляр конфигурации с заданными параметрами.
func New(serverAddr, baseURL, fileStoragePath, databaseDSN, auditFile, auditURL string) Config {
	return Config{
		Server: ServerConfig{
			Address: serverAddr,
			BaseURL: baseURL,
		},
		Storage: StorageConfig{
			FilePath: fileStoragePath,
		},
		Database: DatabaseConfig{
			DSN: databaseDSN,
		},
		Audit: AuditConfig{
			FilePath: auditFile,
			URL:      auditURL,
		},
	}
}

// ParseFlags парсит флаги командной строки и переменные окружения.
// Переменные окружения имеют приоритет над флагами.
// Поддерживаемые флаги:
//
//	-a: адрес сервера (по умолчанию ":8080")
//	-b: базовый URL (по умолчанию "http://localhost:8080")
//	-f: путь к файлу хранилища
//	-d: DSN для PostgreSQL
//	-audit-file: путь к файлу аудита
//	-audit-url: URL удаленного сервера аудита
//
// Поддерживаемые переменные окружения:
//
//	SERVER_ADDRESS, BASE_URL, FILE_STORAGE_PATH, DATABASE_DSN, AUDIT_FILE, AUDIT_URL
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
