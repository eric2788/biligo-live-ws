package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	def := cors.DefaultConfig()
	return cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowWebSockets: true,
		AllowMethods:    def.AllowMethods,
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"Origin",
			"Content-Length",
		},
	})
}
