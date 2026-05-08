package main

import (
	"github.com/esuEdu/investment-risk-engine/internal/config"
	"github.com/esuEdu/investment-risk-engine/internal/db"
	sqlc "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	portfolioHandler "github.com/esuEdu/investment-risk-engine/internal/portfolio/delivery/http"
	"github.com/esuEdu/investment-risk-engine/internal/portfolio/repository"
	"github.com/esuEdu/investment-risk-engine/internal/portfolio/usecase"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/esuEdu/investment-risk-engine/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	logger.Initialize(cfg.AppEnv)
	defer logger.Log.Sync()

	pool, err := db.NewPostgres(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	if err != nil {
		logger.Log.Fatalw("failed to connect to database", "error", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	repo := repository.New(queries, pool)
	uc := usecase.New(repo, cfg.JWTSecret, cfg.APIServiceURL)
	handler := portfolioHandler.New(uc, cfg.JWTSecret)

	router := gin.New()
	router.Use(middleware.ZapLogger())
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"service": "portfolio", "status": "ok"})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
		}

		portfolios := api.Group("/portfolios")
		portfolios.Use(portfolioHandler.AuthMiddleware(cfg.JWTSecret))
		{
			portfolios.POST("", handler.Create)
			portfolios.GET("", handler.List)
			portfolios.GET("/:id", handler.Get)
			portfolios.PUT("/:id", handler.Update)
			portfolios.DELETE("/:id", handler.Delete)
			portfolios.POST("/:id/analyses", handler.TriggerAnalysis)
		}
	}

	logger.Log.Infow("starting portfolio service", "port", "8082")
	if err := router.Run(":8082"); err != nil {
		logger.Log.Fatalw("server failed", "error", err)
	}
}
