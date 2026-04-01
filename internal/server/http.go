package server

import (
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
}

func New() *Server {
	router := gin.New()

	router.Use(ZapLogger())

	router.Use(gin.Recovery())

	s := &Server{
		router: router,
	}

	s.setupRouter()

	return s
}

func (s *Server) Start(addr string) {
	if err := s.router.Run(addr); err != nil {
		logger.Log.Fatalw("Server failed to start", "addr", addr, "error", err)
	}
}
