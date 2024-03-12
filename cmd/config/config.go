// Пакет config отвечает за сбор конфига используя
// глобальные переменные или флаги.
package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"

	"github.com/joho/godotenv"
	"github.com/knstch/shortener/internal/app/logger"
)

// Config хранит важные данные для работы сервера.
type Config struct {
	ServerAddr    string `json:"server_address"`
	BaseURL       string `json:"base_url"`
	FileStorage   string `json:"file_storage_path"`
	DSN           string `json:"database_dsn"`
	SecretKey     string
	EnableHTTPS   bool `json:"enable_https"`
	CertFilePath  string
	KeyFilePath   string
	TrustedSubnet string `json:"trusted_subnet"`
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
	// DSN postgres://admin:password@localhost:7070/?sslmode=disable
	flag.StringVar(&ReadyConfig.DSN, "d", "", "DSN to access DB")
	flag.StringVar(&ReadyConfig.TrustedSubnet, "t", "", "trusted subnet address")
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
	if enableHTTPS := os.Getenv("ENABLE_HTTPS"); enableHTTPS != "" || ReadyConfig.EnableHTTPS {
		ReadyConfig.EnableHTTPS = true
	}
	if trustedSubnet := os.Getenv("TRUSTED_SUBNET"); trustedSubnet != "" {
		ReadyConfig.TrustedSubnet = trustedSubnet
	}
	ReadyConfig.CertFilePath = os.Getenv("CERT_FILE")
	ReadyConfig.KeyFilePath = os.Getenv("PRIVATE_KEY")
	if err = readConfigJSON(); err != nil {
		var pathError *os.PathError
		if errors.As(err, &pathError) {
			return
		} else {
			logger.ErrorLogger("Error reading JSON config: ", err)
		}
	}
}

func readConfigJSON() error {
	f, err := os.Open("../config/config.json")
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, r)
	if err != nil {
		logger.ErrorLogger("Error preparing bytes slice: ", err)
		return err
	}

	err = json.Unmarshal(buffer.Bytes(), &ReadyConfig)
	if err != nil {
		logger.ErrorLogger("Error umrashalling data: ", err)
		return err
	}
	return nil
}
