package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/esuEdu/investment-risk-engine/internal/config"
	"github.com/esuEdu/investment-risk-engine/internal/db"
	sqlc "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	mdHandler "github.com/esuEdu/investment-risk-engine/internal/marketdata/delivery/http"
	"github.com/esuEdu/investment-risk-engine/internal/marketdata/provider"
	"github.com/esuEdu/investment-risk-engine/internal/marketdata/repository"
	"github.com/esuEdu/investment-risk-engine/internal/marketdata/usecase"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/esuEdu/investment-risk-engine/pkg/middleware"
	"github.com/gin-gonic/gin"
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

	queries := sqlc.New(pool)
	repo := repository.New(queries)
	prov := provider.New(cfg.MarketDataAPIKey)
	uc := usecase.New(repo, prov)
	h := mdHandler.New(uc)

	r := gin.New()
	r.Use(middleware.ZapLogger())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	api.GET("/prices", h.GetPrices)
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "marketdata"})
	})

	_ = ctx
	logger.Log.Infow("Market Data Service started", "port", "8081")
	if err := r.Run(":8081"); err != nil {
		logger.Log.Fatalw("server failed", "error", err)
	}
}
