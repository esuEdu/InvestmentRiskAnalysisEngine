package main

import (
	"github.com/esuEdu/investment-risk-engine/internal/config"
	"github.com/esuEdu/investment-risk-engine/internal/db"
	"github.com/esuEdu/investment-risk-engine/internal/server"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
)

func main() {

	cfg := config.Load()

	logger.Initialize(cfg.AppEnv)
	defer logger.Log.Sync()

	logger.Log.Infow("Config and Logger loaded",
		"env", cfg.AppEnv,
		"database", cfg.DBName,
	)

	_, err := db.NewPostgres(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		logger.Log.Fatalw("Failed to connect to database", "error", err)
	}

	s := server.New()
	logger.Log.Info("Starting server on port :8080")
	s.Start(":8080")
}
