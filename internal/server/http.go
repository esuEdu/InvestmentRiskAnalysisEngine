package server

import (
	analyses "github.com/esuEdu/investment-risk-engine/internal/analysis/delivery/http"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router          *gin.Engine
	analysisHandler *analyses.AnalysisHandler
}

func New(analysisHandler *analyses.AnalysisHandler) *Server {
	router := gin.New()

	router.Use(ZapLogger())
	router.Use(gin.Recovery())

	s := &Server{
		router:          router,
		analysisHandler: analysisHandler,
	}

	s.setupRouter()

	return s
}

func (s *Server) Start(addr string) {
	if err := s.router.Run(addr); err != nil {
		logger.Log.Fatalw("Server failed to start", "addr", addr, "error", err)
	}
}
