package app

import "os"

type Config struct {
	HTTPAddress    string
	DatabaseURL    string
	PaymentBaseURL string
}

func LoadConfig() Config {
	return Config{
		HTTPAddress:    getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:55433/order_service?sslmode=disable"),
		PaymentBaseURL: getEnv("PAYMENT_BASE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
