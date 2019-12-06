package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	data := gin.H{
		"status": "OK",
	}
	c.JSONP(http.StatusOK, data)
}
