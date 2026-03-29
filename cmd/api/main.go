package main

import "github.com/esuEdu/investment-risk-engine/internal/server"

func main() {
	s := server.New()
	s.Start(":8080")
}
