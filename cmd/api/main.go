package main

import (
	"log"

	"github.com/esuEdu/investment-risk-engine/internal/config"
	"github.com/esuEdu/investment-risk-engine/internal/db"
	"github.com/esuEdu/investment-risk-engine/internal/server"
)

func main() {

	cfg := config.Load()

	_, err := db.NewPostgres(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatal(err)
	}

	s := server.New()
	s.Start(":8080")
}
