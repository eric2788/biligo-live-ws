package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type appError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err appError) Error() string {
	return fmt.Sprintf("%v: %q", err.Code, err.Message)
}

func ErrorHandler(c *gin.Context) {
	c.Next()
	detectedErrors := c.Errors.ByType(gin.ErrorTypeAny)

	if len(detectedErrors) > 0 {
		err := detectedErrors[0].Err
		log.Infof("Resolving Error: %T", err)
		log.Print(err)
		var parsedError *appError
		switch err.(type) {
		case *appError:
			parsedError = err.(*appError)
		default:
			parsedError = &appError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		c.IndentedJSON(parsedError.Code, parsedError)
		c.Abort()
		return
	}
}
