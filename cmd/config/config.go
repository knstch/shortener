package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr  string
	BaseURL     string
	FileStorage string
	DSN         string
}

var ReadyConfig Config

// Получаем конфиг из флагов, или глобальных переменных, или значения по-умолчанию
func ParseConfig() {
	flag.StringVar(&ReadyConfig.ServerAddr, "a", ":8080", "port to run server")
	flag.StringVar(&ReadyConfig.BaseURL, "b", "http://localhost"+ReadyConfig.ServerAddr, "address to run server")
	flag.StringVar(&ReadyConfig.FileStorage, "f", "short-url-db.json", "file to save links")
	// DSN host=localhost user=postgres password=Xer@0101 dbname=shorten_URLs sslmode=disable
	flag.StringVar(&ReadyConfig.DSN, "d", "", "DSN to access DB")
	flag.Parse()
	if serverAddr := os.Getenv("SERVER_ADDRESS"); serverAddr != "" {
		ReadyConfig.ServerAddr = serverAddr
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		ReadyConfig.BaseURL = baseURL
	}
	if fileStorage := os.Getenv("FILE_STORAGE_PATH"); fileStorage != "" {
		ReadyConfig.FileStorage = fileStorage
	}
	if DSN := os.Getenv("DATABASE_DSN"); DSN != "" {
		ReadyConfig.DSN = DSN
	}
}
