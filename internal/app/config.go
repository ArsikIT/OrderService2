package app

import "os"

type Config struct {
	HTTPAddress     string
	GRPCAddress     string
	DatabaseURL     string
	PaymentGRPCAddr string
}

func LoadConfig() Config {
	return Config{
		HTTPAddress:     getEnv("HTTP_ADDRESS", ":8080"),
		GRPCAddress:     getEnv("GRPC_ADDRESS", ":50052"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:55433/order_service?sslmode=disable"),
		PaymentGRPCAddr: getEnv("PAYMENT_GRPC_ADDR", "127.0.0.1:50051"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
