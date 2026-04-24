package main

import (
	"context"

	analysesHandler "github.com/esuEdu/investment-risk-engine/internal/analysis/delivery/http"
	"github.com/esuEdu/investment-risk-engine/internal/analysis/repository"
	"github.com/esuEdu/investment-risk-engine/internal/analysis/usecase"
	"github.com/esuEdu/investment-risk-engine/internal/config"
	"github.com/esuEdu/investment-risk-engine/internal/db"
	sqlc "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/esuEdu/investment-risk-engine/internal/messaging"
	"github.com/esuEdu/investment-risk-engine/internal/messaging/publisher"
	"github.com/esuEdu/investment-risk-engine/internal/server"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
)

func main() {
	cfg := config.Load()
	logger.Initialize(cfg.AppEnv)
	defer logger.Log.Sync()

	pool, err := db.NewPostgres(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	if err != nil {
		logger.Log.Fatalw("Failed to connect to database", "error", err)
	}
	defer pool.Close()

	ctx := context.Background()

	mq, err := messaging.NewRabbitMQ(ctx, cfg.MQHost, cfg.MQPort, cfg.MQUser, cfg.MQPassword)
	if err != nil {
		logger.Log.Fatalw("Failed to connect to RabbitMQ", "error", err)
	}
	defer mq.Close(ctx)

	pub := publisher.NewPublisher(mq.Conn)
	analysisPublisher := publisher.NewAnalysisPublisher(pub)

	queries := sqlc.New(pool)

	repo := repository.New(queries)
	uc := usecase.New(repo, analysisPublisher)
	handler := analysesHandler.New(uc)

	s := server.New(handler)

	logger.Log.Infow("Starting server", "port", "8080")
	s.Start(":8080")
}
