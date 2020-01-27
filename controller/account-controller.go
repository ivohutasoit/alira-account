package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/ivohutasoit/alira/util"
)

type AccountController struct{}

func (ctrl *AccountController) CreateHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, "account-create.tmpl.html", domain.Page)
	}

	type Request struct {
		Username  string `form:"username" json:"username" xml:"username"`
		Email     string `form:"email" json:"email" xml:"email" binding:"required"`
		Mobile    string `form:"mobile" json:"mobile" xml:"mobile" binding:"required"`
		FirstName string `form:"first_name" json:"first_name" xml:"first_name" binding:"required"`
		LastName  string `form:"last_name" json:"last_name" xml:"last_name" binding:"required"`
		Active    bool   `form:"active" json:"active" xml:"active"`
	}

	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))
	var req Request
	if api {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
		}
	}
	acctService := &service.AccountService{}
	data, err := acctService.Create(req.Username, req.Email, req.Mobile, req.FirstName, req.LastName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":   http.StatusBadRequest,
			"status": http.StatusText(http.StatusBadRequest),
			"error":  err.Error(),
		})
	}
	user := data["user"].(*domain.User)
	if data["status"].(string) == "SUCCESS" {
		if api {
			c.JSON(http.StatusCreated, gin.H{
				"code":    http.StatusCreated,
				"status":  http.StatusText(http.StatusCreated),
				"message": "Profile has been created",
				"data": map[string]string{
					"user_id": user.BaseModel.ID,
				},
			})
		}
	}
}

func (ctrl *AccountController) DetailHandler(c *gin.Context) {
	var id string
	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))
	if api {
		id = c.Param("id")
	} else {
		id = c.Query("id")
	}

	accountService := &service.AccountService{}
	data, err := accountService.Get(id)
	if err != nil {
		if api {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
		}
		return
	}

	user := data["user"].(*domain.User)
	profile := data["profile"].(*domain.Profile)
	fmt.Println(user.Username)
	if api {
		c.JSON(http.StatusOK, gin.H{
			"code":   http.StatusOK,
			"status": http.StatusText(http.StatusOK),
			"data": map[string]string{
				"username": user.Username,
				"email":    user.Email,
				"mobile":   user.Mobile,
				"name":     fmt.Sprintf("%s %s", profile.FirstName, profile.LastName),
			},
		})
	}
}

func AccountViewHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		return
	}
}

func RegisterHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, nil)
		return
	}

	type Request struct {
		Reference string `form:"reference" json:"reference" xml:"reference" binding:"required"`
		UserID    string `form:"userid" json:"userid" xml:"userid" binding:"required"`
	}

	var request Request
	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	} else {
		if err := c.ShouldBind(&request); err != nil {
			c.HTML(http.StatusBadRequest, constant.RegisterPage, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.Set("userid", request.UserID)

	switch request.Reference {
	case "mobile":
		RegisterByMobileHandler(c)
		return
	default:
		RegisterByEmailHandler(c)
		return
	}
}

// RegisterByEmailHandler for user registration using email address
func RegisterByEmailHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, nil)
		return
	}

	accountService := &service.AccountService{}
	data, err := accountService.SendRegisterToken(c.GetString("userid"), "email")
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
		} else {
			c.HTML(http.StatusOK, constant.RegisterPage, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	if data["status"].(string) == "SUCCESS" {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.JSONP(http.StatusOK, gin.H{
				"code":   200,
				"status": "OK",
				"data": map[string]string{
					"referer": data["referer"].(string),
					"next":    "activate",
				},
			})
			return
		}

		c.HTML(http.StatusOK, constant.TokenPage, gin.H{
			"referer": data["referer"].(string),
			"purpose": data["purpose"].(string),
			"message": data["message"].(string),
		})
	}
}

func RegisterByMobileHandler(c *gin.Context) {
	c.JSONP(http.StatusOK, gin.H{
		"code":   200,
		"status": "OK",
	})
}

func ProfileHandler(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		action = "view"
	}
	accService := &service.AccountService{}
	data, err := accService.Get(c.GetString("user_id"))
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
		} else {
			c.HTML(http.StatusBadRequest, constant.ProfilePage, gin.H{
				"user_id": c.GetString("user_id"),
				"error":   err.Error(),
			})
		}
		return
	}

	user := data["user"].(*domain.User)
	profile := data["profile"].(*domain.Profile)
	if c.Request.Method == http.MethodGet {
		if user.Username == "" {
			action = "complete"
		}
		c.HTML(http.StatusOK, constant.ProfilePage, gin.H{
			"user_id":    user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"mobile":     user.Mobile,
			"first_name": profile.FirstName,
			"last_name":  profile.LastName,
			"gender":     strings.ToLower(profile.Gender),
			"state":      action,
		})
		return
	}
	type Request struct {
		Username  string `form:"username" json:"username" xml:"username" binding:"required"`
		Mobile    string `form:"mobile" json:"mobile" xml:"mobile" binding:"required"`
		FirstName string `form:"first_name" json:"first_name" xml:"first_name" binding:"required"`
		LastName  string `form:"last_name" json:"last_name" xml:"last_name" binding:"required"`
		Gender    string `form:"gender" json:"gender" xml:"gender" binding:"required"`
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
			c.HTML(http.StatusBadRequest, constant.ProfilePage, gin.H{
				"user_id":    user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"mobile":     user.Mobile,
				"first_name": profile.FirstName,
				"last_name":  profile.LastName,
				"gender":     strings.ToLower(profile.Gender),
				"state":      action,
				"error":      err.Error(),
			})
			return
		}
	}
	data, err = accService.SaveProfile(c.GetString("user_id"), req.Username, req.Mobile, req.FirstName, req.LastName, req.Gender)
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
		} else {
			c.HTML(http.StatusBadRequest, constant.ProfilePage, gin.H{
				"user_id":    user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"mobile":     user.Mobile,
				"first_name": profile.FirstName,
				"last_name":  profile.LastName,
				"gender":     strings.ToLower(profile.Gender),
				"state":      action,
				"error":      err.Error(),
			})
		}
		return
	}

	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"status":  "OK",
			"message": data["message"].(string),
			"data": map[string]string{
				"user_id": data["user_id"].(string),
			},
		})
		return
	}

	session := sessions.Default(c)
	session.Set("message", data["message"])
	session.Save()
	uri, _ := util.GenerateUrl(c.Request.TLS, c.Request.Host, "/account/profile", false)
	c.Redirect(http.StatusMovedPermanently, uri)
}
