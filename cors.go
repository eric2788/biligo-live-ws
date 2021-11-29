package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowWebSockets: true,
		AllowHeaders: []string{
			"X-BLive-Identifier",
			"Content-Type",
		},
	})
}
