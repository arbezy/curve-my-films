package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

// TODO: at some point will need to widen this to all app config rather than just db conf vals
type DBConfig struct {
	DB_USERNAME string
	DB_PASSWORD string
	DB_NET      string
	DB_ADDRESS  string
	DB_NAME     string
}

var DatabaseConfig DBConfig

func Init() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	DatabaseConfig = DBConfig{
		DB_USERNAME: getEnv("DB_USERNAME"),
		DB_PASSWORD: getEnv("DB_PASSWORD"),
		DB_NET:      getEnv("DB_NET"),
		DB_ADDRESS:  getEnv("DB_ADDRESS"),
		DB_NAME:     getEnv("DB_NAME"),
	}

	return nil
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Missing required env var: %s", key))
	}
	return val
}
