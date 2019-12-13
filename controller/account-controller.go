package controller

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
	srvc "github.com/ivohutasoit/alira/service"
	"github.com/ivohutasoit/alira/util"
)

func RegisterHandler(c *gin.Context) {
	step := c.Query("step")
	by := c.Query("by")

	switch step {
	case "activate":
		ActivateTokenHandler(c)
		return
	case "profile":
		CompleteProfileHandler(c)
		return
	case "upgrade":
		return
	default:
		switch by {
		case "mobile":
			RegisterByMobileHandler(c)
			return
		default:
			RegisterByEmailHandler(c)
			return
		}
	}
}

// RegisterByEmailHandler for user registration using email address
func RegisterByEmailHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, nil)
	} else {
		type Registration struct {
			Email string `form:"email" json:"email" xml:"email"  binding:"required"`
		}
		var registration Registration
		if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
			if err := c.ShouldBindJSON(&registration); err != nil {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusBadRequest, gin.H{
					"code":   400,
					"status": "Bad Request",
					"error":  err.Error(),
				})
				return
			}
		} else {
			if err := c.ShouldBind(&registration); err != nil {
				c.HTML(http.StatusBadRequest, constant.RegisterPage, nil)
				return
			}
		}

		token := &domain.Token{
			BaseModel: model.BaseModel{
				ID: uuid.New().String(),
			},
			UserID:    registration.Email,
			Code:      util.GenerateToken(6),
			Purpose:   "REGISTRATION",
			ExpiredAt: time.Now().Add(time.Hour * 12),
			Valid:     true,
		}
		model.GetDatabase().Create(token)

		mail := &domain.Mail{
			From:     os.Getenv("SMTP.SENDER"),
			To:       []string{registration.Email},
			Subject:  "Token Registration",
			Template: "templates/mail/token_registration.html",
			Data: map[interface{}]interface{}{
				"token": token.Code,
			},
		}
		ms := &srvc.MailService{}
		result, err := ms.Send(mail)
		if err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}

		c.JSONP(http.StatusOK, gin.H{
			"code":    200,
			"status":  "OK",
			"message": result["message"].(string),
			"data": map[string]string{
				"referer": registration.Email,
			},
		})
	}
}

func RegisterByMobileHandler(c *gin.Context) {
	c.JSONP(http.StatusOK, gin.H{
		"code":   200,
		"status": "OK",
	})
}

func ActivateTokenHandler(c *gin.Context) {
	c.JSONP(http.StatusOK, gin.H{
		"code":   200,
		"status": "OK",
	})
	type RegistrationToken struct {
		Referer string `form:"referer" json:"referer" xml:"referer" binding:"required"` // email, mobile
		Code    string `form:"code" json:"code" xml:"code"  binding:"required"`
	}
	var registerToken RegistrationToken
	if err := c.ShouldBindJSON(&registerToken); err != nil {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  err.Error(),
		})
		return
	}

	var token *domain.Token
	model.GetDatabase().First(&token, "user_id = ? AND code = ?",
		registerToken.Referer, registerToken.Code)

	if token == nil {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  "invalid token",
		})
		return
	}

	if token.ExpiredAt.Before(time.Now()) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  "invalid token",
		})
		return
	}

	if !token.Valid {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  "invalid token",
		})
		return
	}

	model.GetDatabase().Delete(&token)

	c.JSONP(http.StatusOK, gin.H{
		"code":    200,
		"status":  "OK",
		"message": "token has been activated successfully",
		"data": map[string]string{
			"referer": registerToken.Referer,
		},
	})
}

func CompleteProfileHandler(c *gin.Context) {
	type Profile struct {
		Username  string `form:"username" json:"username" xml:"username"  binding:"required"`
		Email     string `form:"email" json:"email" xml:"email"  binding:"required"`
		Mobile    string `form:"mobile" json:"mobile" xml:"mobile"  binding:"required"`
		FirstName string `form:"first_name" json:"first_name" xml:"first_name"  binding:"required"`
		LastName  string `form:"last_name" json:"last_name" xml:"last_name"  binding:"required"`
		Package   string `form:"package" json:"package" xml:"package"  binding:"required"`
	}
	var profile Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  err.Error(),
		})
		return
	}

}
