package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func Identifier() gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.GetHeader("Authorization")
		if identifier == "" {
			identifier = "anonymous"
		}
		c.Set("identifier", fmt.Sprintf("%v@%v", c.ClientIP(), identifier))
		c.Next()
	}
}
