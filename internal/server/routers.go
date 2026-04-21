package server

import "github.com/gin-gonic/gin"

func (s *Server) setupRouter() {

	api := s.router.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)
		// Analysis Routes
		analysis := api.Group("/analyses")
		{
			analysis.POST("", s.analysisHandler.Create)
			analysis.GET("", s.analysisHandler.Get)
			analysis.GET("", s.analysisHandler.List)
			analysis.PUT("", s.analysisHandler.Update)
		}
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	OK(c, "server is running", gin.H{
		"service": "api",
	})
}
