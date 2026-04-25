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
}

func Load() *Config {
	viper.SetConfigFile(".env")

	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No .env file found, using environment variables or defaults")
		} else {
			log.Fatalf("Fatal error reading config file: %s \n", err)
		}
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
	}
}
