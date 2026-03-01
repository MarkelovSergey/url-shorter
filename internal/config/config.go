// Package config содержит конфигурацию приложения.
package config

import (
	"encoding/json"
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
	enableHTTPSEnv     = "ENABLE_HTTPS"
	configFileEnv      = "CONFIG"
	trustedSubnetEnv   = "TRUSTED_SUBNET"
	grpcAddressEnv     = "GRPC_ADDRESS"
)

// JSONConfig представляет структуру JSON файла конфигурации.
type JSONConfig struct {
	// ServerAddress - адрес HTTP-сервера (аналог -a или SERVER_ADDRESS)
	ServerAddress string `json:"server_address,omitempty"`
	// BaseURL - базовый URL для создания коротких ссылок (аналог -b или BASE_URL)
	BaseURL string `json:"base_url,omitempty"`
	// FileStoragePath - путь к файлу для хранения URL (аналог -f или FILE_STORAGE_PATH)
	FileStoragePath string `json:"file_storage_path,omitempty"`
	// DatabaseDSN - строка подключения к PostgreSQL (аналог -d или DATABASE_DSN)
	DatabaseDSN string `json:"database_dsn,omitempty"`
	// EnableHTTPS - включает HTTPS-сервер вместо HTTP (аналог -s или ENABLE_HTTPS)
	EnableHTTPS *bool `json:"enable_https,omitempty"`
	// AuditFile - путь к файлу для записи событий аудита (аналог -audit-file или AUDIT_FILE)
	AuditFile string `json:"audit_file,omitempty"`
	// AuditURL - URL удаленного сервера для отправки событий аудита (аналог -audit-url или AUDIT_URL)
	AuditURL string `json:"audit_url,omitempty"`
	// TrustedSubnet - доверенная подсеть в нотации CIDR (аналог -t или TRUSTED_SUBNET)
	TrustedSubnet string `json:"trusted_subnet,omitempty"`
	// GRPCAddress - адрес gRPC-сервера (аналог -g или GRPC_ADDRESS)
	GRPCAddress string `json:"grpc_address,omitempty"`
}

// ServerConfig содержит настройки HTTP-сервера.
type ServerConfig struct {
	// Address - адрес HTTP-сервера (например, ":8080" или "localhost:8888")
	Address string
	// BaseURL - базовый URL для создания коротких ссылок (например, "http://localhost:8080")
	BaseURL string
	// EnableHTTPS - включает HTTPS-сервер вместо HTTP
	EnableHTTPS bool
	// TrustedSubnet - доверенная подсеть в нотации CIDR для доступа к /api/internal/stats
	TrustedSubnet string
	// GRPCAddress - адрес gRPC-сервера (например, ":3200")
	GRPCAddress string
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
func New(
	serverAddr string,
	baseURL string,
	fileStoragePath string,
	databaseDSN string,
	auditFile string,
	auditURL string,
	trustedSubnet string,
	grpcAddress string,
	enableHTTPS bool,
) Config {
	return Config{
		Server: ServerConfig{
			Address:       serverAddr,
			BaseURL:       baseURL,
			EnableHTTPS:   enableHTTPS,
			TrustedSubnet: trustedSubnet,
			GRPCAddress:   grpcAddress,
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

// loadJSONConfig загружает конфигурацию из JSON файла.
// Возвращает nil, если файл не существует или произошла ошибка.
func loadJSONConfig(filePath string) *JSONConfig {
	if filePath == "" {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		// Файл не существует или не может быть прочитан - это нормально
		return nil
	}

	var cfg JSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Ошибка парсинга JSON - игнорируем
		return nil
	}

	return &cfg
}

// flagValues содержит значения всех флагов конфигурации.
type flagValues struct {
	serverAddr      string
	baseURL         string
	fileStoragePath string
	databaseDSN     string
	auditFile       string
	auditURL        string
	trustedSubnet   string
	grpcAddress     string
	enableHTTPS     bool
	configFile      string
}

// defineFlags определяет и парсит флаги командной строки.
// Возвращает структуру с значениями по умолчанию из флагов.
func defineFlags() *flagValues {
	serverAddr := flag.String("a", ":8080", "HTTP server address (e.g. localhost:8888)")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	configFile := flag.String("c", "", "path to JSON config file")
	databaseDSN := flag.String("d", "", "database connection string")
	fileStoragePath := flag.String("f", "/var/lib/url-shorter/short-url-db.json", "file storage path")
	enableHTTPS := flag.Bool("s", false, "enable HTTPS server")
	trustedSubnet := flag.String("t", "", "trusted subnet in CIDR notation for /api/internal/stats")
	grpcAddress := flag.String("g", ":3200", "gRPC server address")
	auditFile := flag.String("audit-file", "", "path to audit log file")
	auditURL := flag.String("audit-url", "", "URL of remote audit server")
	flag.StringVar(configFile, "config", "", "path to JSON config file")
	flag.Parse()

	return &flagValues{
		serverAddr:      *serverAddr,
		baseURL:         *baseURL,
		fileStoragePath: *fileStoragePath,
		databaseDSN:     *databaseDSN,
		auditFile:       *auditFile,
		auditURL:        *auditURL,
		trustedSubnet:   *trustedSubnet,
		grpcAddress:     *grpcAddress,
		enableHTTPS:     *enableHTTPS,
		configFile:      *configFile,
	}
}

// getConfigFilePath определяет путь к конфигурационному файлу.
// Переменная окружения CONFIG имеет приоритет над флагом.
func getConfigFilePath(flagValue string) string {
	if envConfigPath, ok := os.LookupEnv(configFileEnv); ok {
		return envConfigPath
	}
	return flagValue
}

// applyJSONConfig применяет значения из JSON конфигурации к flagValues.
// Если поле в JSON не задано, значение из values не изменяется.
func applyJSONConfig(values *flagValues, jsonCfg *JSONConfig) {
	if jsonCfg == nil {
		return
	}

	if jsonCfg.ServerAddress != "" {
		values.serverAddr = jsonCfg.ServerAddress
	}
	if jsonCfg.BaseURL != "" {
		values.baseURL = jsonCfg.BaseURL
	}
	if jsonCfg.FileStoragePath != "" {
		values.fileStoragePath = jsonCfg.FileStoragePath
	}
	if jsonCfg.DatabaseDSN != "" {
		values.databaseDSN = jsonCfg.DatabaseDSN
	}
	if jsonCfg.AuditFile != "" {
		values.auditFile = jsonCfg.AuditFile
	}
	if jsonCfg.AuditURL != "" {
		values.auditURL = jsonCfg.AuditURL
	}
	if jsonCfg.EnableHTTPS != nil {
		values.enableHTTPS = *jsonCfg.EnableHTTPS
	}
	if jsonCfg.TrustedSubnet != "" {
		values.trustedSubnet = jsonCfg.TrustedSubnet
	}
	if jsonCfg.GRPCAddress != "" {
		values.grpcAddress = jsonCfg.GRPCAddress
	}
}

// applyEnvVariables применяет значения из переменных окружения к flagValues.
// Переменные окружения имеют наивысший приоритет.
func applyEnvVariables(values *flagValues) {
	if envServerAddr, ok := os.LookupEnv(serverAddressEnv); ok {
		values.serverAddr = envServerAddr
	}

	if envBaseURL, ok := os.LookupEnv(baseURLEnv); ok {
		values.baseURL = envBaseURL
	}

	if envFileStoragePath, ok := os.LookupEnv(fileStoragePathEnv); ok {
		values.fileStoragePath = envFileStoragePath
	}

	if envDatabaseDSN, ok := os.LookupEnv(databaseDSNEnv); ok {
		values.databaseDSN = envDatabaseDSN
	}

	if envAuditFile, ok := os.LookupEnv(auditFileEnv); ok {
		values.auditFile = envAuditFile
	}

	if envAuditURL, ok := os.LookupEnv(auditURLEnv); ok {
		values.auditURL = envAuditURL
	}

	if envEnableHTTPS, ok := os.LookupEnv(enableHTTPSEnv); ok {
		values.enableHTTPS = envEnableHTTPS == "true" || envEnableHTTPS == "1"
	}

	if envTrustedSubnet, ok := os.LookupEnv(trustedSubnetEnv); ok {
		values.trustedSubnet = envTrustedSubnet
	}

	if envGRPCAddress, ok := os.LookupEnv(grpcAddressEnv); ok {
		values.grpcAddress = envGRPCAddress
	}
}

// ParseFlags парсит флаги командной строки и переменные окружения.
// Приоритет (от низкого к высокому):
//  1. Значения по умолчанию флагов
//  2. Значения из JSON файла конфигурации (если указан через -c/-config или CONFIG)
//  3. Переменные окружения
//  4. Флаги командной строки
//
// Поддерживаемые флаги:
//
//	-a: адрес сервера (по умолчанию ":8080")
//	-b: базовый URL (по умолчанию "http://localhost:8080")
//	-f: путь к файлу хранилища
//	-d: : DSN для PostgreSQL
//	-s: включить HTTPS сервер
//	-c / --config: путь к файлу конфигурации в формате JSON
//	-audit-file: путь к файлу аудита
//	-audit-url: URL удаленного сервера аудита
//
// Поддерживаемые переменные окружения:
//
//	SERVER_ADDRESS, BASE_URL, FILE_STORAGE_PATH, DATABASE_DSN, ENABLE_HTTPS, CONFIG, AUDIT_FILE, AUDIT_URL
func ParseFlags() Config {
	// 1. Определяем и парсим флаги командной строки
	values := defineFlags()

	// 2. Определяем путь к конфигурационному файлу
	configPath := getConfigFilePath(values.configFile)

	// 3. Применяем значения из JSON конфигурации (если есть)
	jsonCfg := loadJSONConfig(configPath)
	applyJSONConfig(values, jsonCfg)

	// 4. Применяем значения из переменных окружения (наивысший приоритет)
	applyEnvVariables(values)

	// 5. Создаем и возвращаем итоговую конфигурацию
	return New(
		values.serverAddr,
		values.baseURL,
		values.fileStoragePath,
		values.databaseDSN,
		values.auditFile,
		values.auditURL,
		values.trustedSubnet,
		values.grpcAddress,
		values.enableHTTPS,
	)
}
