package server

import "github.com/gin-gonic/gin"

func (s *Server) setupRouter() {

	api := s.router.Group("/api")
	{
		api.GET("/health", s.healthCheck)
		// Analysis Routes
		analysis := api.Group("/analyses")
		{
			analysis.POST("", s.analysisHandler.Create)
		}
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	OK(c, "server is running", gin.H{
		"service": "api",
	})
}
