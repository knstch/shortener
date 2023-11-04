package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr  string
	BaseURL     string
	FileStorage string
	DBDSN       string
	DBUsername  string
	DBPassword  string
	DBName      string
}

var ReadyConfig Config

// Получаем конфиг из флагов, или глобальных переменных, или значения по-умолчанию
func ParseConfig() {
	flag.StringVar(&ReadyConfig.ServerAddr, "a", ":8080", "port to run server")
	flag.StringVar(&ReadyConfig.BaseURL, "b", "http://localhost"+ReadyConfig.ServerAddr, "address to run server")
	flag.StringVar(&ReadyConfig.FileStorage, "f", "short-url-db.json", "file to save links")
	flag.StringVar(&ReadyConfig.DBDSN, "d", "localhost", "DSN to access DB")
	flag.StringVar(&ReadyConfig.DBUsername, "u", "postgres", "Username for DB access")
	flag.StringVar(&ReadyConfig.DBPassword, "p", "Xer@0101", "Password for DB access")
	flag.StringVar(&ReadyConfig.DBName, "db", "shorten_URLs", "Database name")
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
	if dbDSN := os.Getenv("DATABASE_DSN"); dbDSN != "" {
		ReadyConfig.DBDSN = dbDSN
	}
	if DBName := os.Getenv("POSTGRES_DB"); DBName != "" {
		ReadyConfig.DBName = DBName
	}
	if DBPassword := os.Getenv("POSTGRES_PASSWORD"); DBPassword != "" {
		ReadyConfig.DBPassword = DBPassword
	}
}
