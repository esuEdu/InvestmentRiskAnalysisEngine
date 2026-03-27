package server

import (
	"log"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
}

func New() *Server {
	s := &Server{
		router: gin.Default(),
	}

	s.setupRouter()

	return s
}

func (s *Server) Start(addr string) {
	if err := s.router.Run(addr); err != nil {
		log.Fatal(err)
	}
}
