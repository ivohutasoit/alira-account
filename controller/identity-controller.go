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
		Document      string `form:"document" json:"document" xml:"document" binding:"required"`
		NationID      string `form:"nation_id" json:"nation_id" xml:"nation_id" binding:"required"`
		Address       string `form:"address" json:"address" xml:"address"`
		City          string `form:"city" json:"city" xml:"city"`
		State         string `form:"state" json:"state" xml:"state"`
		Province      string `form:"province" json:"province" xml:"province"`
		Country       string `form:"country" json:"country" xml:"country" binding:"required"`
		PostalCode    string `form:"postal_code" json:"postal_code" xml:"postal_code"`
		BloodType     string `form:"blood_type" json:"blood_type" xml:"blood_type"`
		Religion      string `form:"religion" json:"religion" xml:"religion"`
		MarriedStatus string `form:"married_status" json:"married_status" xml:"married_status"`
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

	if req.Document != "E-KTP" || req.Document != "PASSPORT" {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  "invalid document type",
			})
			return
		} else {
			return
		}
	}

}
