package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/esuEdu/investment-risk-engine/internal/analysis/repository"
	"github.com/esuEdu/investment-risk-engine/internal/config"
	"github.com/esuEdu/investment-risk-engine/internal/db"
	sqlc "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/esuEdu/investment-risk-engine/internal/messaging"
	"github.com/esuEdu/investment-risk-engine/internal/messaging/consumer"
	"github.com/esuEdu/investment-risk-engine/internal/riskmetrics"
	"github.com/esuEdu/investment-risk-engine/internal/worker"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger.Initialize(cfg.AppEnv)
	defer logger.Log.Sync()

	pool, err := db.NewPostgres(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		logger.Log.Fatalw("failed to connect to database", "error", err)
	}
	defer pool.Close()

	mq, err := messaging.NewRabbitMQ(ctx, cfg.MQHost, cfg.MQPort, cfg.MQUser, cfg.MQPassword)
	if err != nil {
		logger.Log.Fatalw("failed to connect to RabbitMQ", "error", err)
	}
	defer mq.Close(ctx)

	queries := sqlc.New(pool)
	repo := repository.New(queries, pool)

	mdClient := worker.NewHTTPMarketDataClient(cfg.MarketDataServiceURL)
	calculator := riskmetrics.NewCalculator()
	workerHandler := worker.New(repo, mdClient, calculator)

	analysisConsumer := consumer.NewAnalysisConsumer(consumer.NewConsumer(mq.Conn))

	logger.Log.Infow("Worker Service started — listening on risk-analysis-jobs")

	if err := analysisConsumer.Start(ctx, func(req *domain.AnalysisRequest) error {
		return workerHandler.Handle(ctx, req)
	}); err != nil {
		logger.Log.Fatalw("consumer stopped with error", "error", err)
	}
}
