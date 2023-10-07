package config

import (
	"flag"
	"os"
)

var BasicAddr string
var Port string

type Config struct {
	serverAddr string `env:"SERVER_ADDRESS"`
	baseURL    string `env:"BASE_URL"`
}

// Получаем конфиг из флагов, или глобальных переменных, или значения по-умолчанию
func ParseENV() {
	flag.StringVar(&Port, "a", ":8080", "port to run server")
	flag.StringVar(&BasicAddr, "b", "http://localhost"+Port, "address to run server")
	flag.Parse()

	var cfg Config
	cfg.serverAddr = os.Getenv("SERVER_ADDRESS")
	cfg.baseURL = os.Getenv("BASE_URL")
	if cfg.serverAddr != "" {
		Port = cfg.serverAddr
	}
	if cfg.baseURL != "" {
		BasicAddr = cfg.baseURL
	}
}
