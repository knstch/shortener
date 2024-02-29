// Пакет config отвечает за сбор конфига используя
// глобальные переменные или флаги.
package config

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/knstch/shortener/internal/app/logger"
)

// Config хранит важные данные для работы сервера.
type Config struct {
	ServerAddr   string
	BaseURL      string
	FileStorage  string
	DSN          string
	SecretKey    string
	EnableHTTPS  bool
	CertFilePath string
	KeyFilePath  string
}

// ReadyConfig хранит config.
var ReadyConfig Config

// ParseConfig собирает config параметры из флагов и переменных окружения.
func ParseConfig() {
	err := godotenv.Load("../../.env")
	if err != nil {
		logger.ErrorLogger("Error parsing .env: ", err)
	}
	flag.StringVar(&ReadyConfig.ServerAddr, "a", ":8080", "port to run server")
	flag.StringVar(&ReadyConfig.BaseURL, "b", "http://localhost"+ReadyConfig.ServerAddr, "address to run server")
	flag.StringVar(&ReadyConfig.FileStorage, "f", "short-url-db.json", "file to save links")
	// DSN postgres://postgres:Xer_0101@localhost:5432/shorten_urls?sslmode=disable
	flag.StringVar(&ReadyConfig.DSN, "d", "", "DSN to access DB")
	flag.BoolVar(&ReadyConfig.EnableHTTPS, "s", false, "enabling HTTPS connection")
	flag.Parse()
	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		ReadyConfig.SecretKey = secretKey
	}
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
	if enableHTTPS := os.Getenv("ENABLE_HTTPS"); enableHTTPS != "" {
		ReadyConfig.EnableHTTPS = true
	}
	ReadyConfig.CertFilePath = os.Getenv("CERT_FILE")
	ReadyConfig.KeyFilePath = os.Getenv("PRIVATE_KEY")
}
