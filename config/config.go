package config

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
