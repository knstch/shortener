package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

var ReadyConfig Config

// Получаем конфиг из флагов, или глобальных переменных, или значения по-умолчанию
func ParseConfig() {
	flag.StringVar(&ReadyConfig.ServerAddr, "a", ":8080", "port to run server")
	flag.StringVar(&ReadyConfig.BaseURL, "b", "http://localhost"+ReadyConfig.ServerAddr, "address to run server")

	flag.Parse()

	if serverAddr := os.Getenv("SERVER_ADDRESS"); serverAddr != "" {
		ReadyConfig.ServerAddr = serverAddr
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		ReadyConfig.BaseURL = baseURL
	}
}
