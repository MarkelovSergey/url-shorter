package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoadJSONConfig(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantNil  bool
		wantAddr string
	}{
		{
			name:     "valid config",
			content:  `{"server_address": "localhost:9090", "base_url": "http://test.com"}`,
			wantNil:  false,
			wantAddr: "localhost:9090",
		},
		{
			name:    "invalid json",
			content: `{invalid json}`,
			wantNil: true,
		},
		{
			name:    "empty file",
			content: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл
			tmpFile, err := os.CreateTemp("", "config-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			if tt.content != "" {
				if _, err := tmpFile.WriteString(tt.content); err != nil {
					t.Fatal(err)
				}
			}
			tmpFile.Close()

			cfg := loadJSONConfig(tmpFile.Name())

			if tt.wantNil && cfg != nil {
				t.Errorf("expected nil, got %+v", cfg)
			}
			if !tt.wantNil && cfg == nil {
				t.Error("expected non-nil config")
			}
			if !tt.wantNil && cfg != nil && cfg.ServerAddress != tt.wantAddr {
				t.Errorf("expected server_address %q, got %q", tt.wantAddr, cfg.ServerAddress)
			}
		})
	}
}

func TestLoadJSONConfigNotExists(t *testing.T) {
	cfg := loadJSONConfig("/nonexistent/file.json")
	if cfg != nil {
		t.Errorf("expected nil for non-existent file, got %+v", cfg)
	}
}

func TestLoadJSONConfigEmpty(t *testing.T) {
	cfg := loadJSONConfig("")
	if cfg != nil {
		t.Errorf("expected nil for empty path, got %+v", cfg)
	}
}

func TestJSONConfigPriority(t *testing.T) {
	// Сохраняем оригинальные значения переменных окружения
	origServerAddr := os.Getenv(serverAddressEnv)
	origBaseURL := os.Getenv(baseURLEnv)
	origConfigFile := os.Getenv(configFileEnv)
	defer func() {
		os.Setenv(serverAddressEnv, origServerAddr)
		os.Setenv(baseURLEnv, origBaseURL)
		os.Setenv(configFileEnv, origConfigFile)
	}()

	// Очищаем переменные окружения
	os.Unsetenv(serverAddressEnv)
	os.Unsetenv(baseURLEnv)
	os.Unsetenv(configFileEnv)

	// Создаем JSON конфигурацию
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	jsonContent := `{
		"server_address": "localhost:9090",
		"base_url": "http://json.local",
		"file_storage_path": "/json/path.db",
		"enable_https": true
	}`
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Тест 1: JSON значения применяются поверх значений по умолчанию
	t.Run("json overrides defaults", func(t *testing.T) {
		// Сбрасываем флаги для каждого теста
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"cmd", "-c", tmpFile.Name()}

		cfg := ParseFlags()

		if cfg.Server.Address != "localhost:9090" {
			t.Errorf("expected server address from JSON, got %q", cfg.Server.Address)
		}
		if cfg.Server.BaseURL != "http://json.local" {
			t.Errorf("expected base URL from JSON, got %q", cfg.Server.BaseURL)
		}
		if cfg.Storage.FilePath != "/json/path.db" {
			t.Errorf("expected file path from JSON, got %q", cfg.Storage.FilePath)
		}
		if !cfg.Server.EnableHTTPS {
			t.Error("expected enable_https to be true from JSON")
		}
	})

	// Тест 2: Переменные окружения имеют приоритет над JSON
	t.Run("env overrides json", func(t *testing.T) {
		os.Setenv(serverAddressEnv, "env:8888")
		os.Setenv(baseURLEnv, "http://env.local")
		defer func() {
			os.Unsetenv(serverAddressEnv)
			os.Unsetenv(baseURLEnv)
		}()

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"cmd", "-c", tmpFile.Name()}

		cfg := ParseFlags()

		if cfg.Server.Address != "env:8888" {
			t.Errorf("expected server address from env, got %q", cfg.Server.Address)
		}
		if cfg.Server.BaseURL != "http://env.local" {
			t.Errorf("expected base URL from env, got %q", cfg.Server.BaseURL)
		}
		// Значения из JSON должны остаться где нет env переменных
		if cfg.Storage.FilePath != "/json/path.db" {
			t.Errorf("expected file path from JSON, got %q", cfg.Storage.FilePath)
		}
	})

	// Тест 3: CONFIG переменная окружения для указания пути к конфигу
	t.Run("CONFIG env variable", func(t *testing.T) {
		os.Setenv(configFileEnv, tmpFile.Name())
		os.Unsetenv(serverAddressEnv)
		os.Unsetenv(baseURLEnv)
		defer os.Unsetenv(configFileEnv)

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"cmd"}

		cfg := ParseFlags()

		if cfg.Server.Address != "localhost:9090" {
			t.Errorf("expected server address from JSON via CONFIG env, got %q", cfg.Server.Address)
		}
	})
}
