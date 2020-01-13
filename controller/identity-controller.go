package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
)

func IdentityHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.IdentityPage, nil)
		return
	}

	type Request struct {
		NationID   string `form:"nation_id" json:"nation_id" xml:"nation_id" binding:"required"`
		Address    string `form:"address" json:"address" xml:"address" binding:"required"`
		City       string `form:"city" json:"city" xml:"city" binding:"required"`
		State      string `form:"state" json:"state" xml:"state" binding:"required"`
		Province   string `form:"province" json:"province" xml:"provice" binding:"required"`
		Country    string `form:"country" json:"country" xml:"country" binding:"required"`
		PostalCode string `form:"postal_code" json:"postal_code" xml:"postal_code" binding:"required"`
	}

	var req Request
	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			return
		}
	}

	
}
