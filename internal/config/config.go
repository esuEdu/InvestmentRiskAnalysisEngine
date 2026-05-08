package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	MQHost     string
	MQPort     int
	MQUser     string
	MQPassword string

	MarketDataServiceURL string
	MarketDataAPIKey     string

	JWTSecret     string
	APIServiceURL string
}

func Load() *Config {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		AppEnv:     viper.GetString("APP_ENV"),
		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		MQHost:     viper.GetString("MQ_HOST"),
		MQPort:     viper.GetInt("MQ_PORT"),
		MQUser:     viper.GetString("MQ_USER"),
		MQPassword: viper.GetString("MQ_PASSWORD"),

		MarketDataServiceURL: viper.GetString("MARKET_DATA_SERVICE_URL"),
		MarketDataAPIKey:     viper.GetString("MARKET_DATA_API_KEY"),

		JWTSecret:     viper.GetString("JWT_SECRET"),
		APIServiceURL: viper.GetString("API_SERVICE_URL"),
	}
}
