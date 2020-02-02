package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/ivohutasoit/alira/messaging"
	"github.com/ivohutasoit/alira/util"
)

type Account struct{}

func (ctrl *Account) CreateHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, "account-create.tmpl.html", alira.ViewData)
	}

	type Request struct {
		Username     string `form:"username" json:"username" xml:"username"`
		Email        string `form:"email" json:"email" xml:"email" binding:"required"`
		Mobile       string `form:"mobile" json:"mobile" xml:"mobile" binding:"required"`
		FirstName    string `form:"first_name" json:"first_name" xml:"first_name" binding:"required"`
		LastName     string `form:"last_name" json:"last_name" xml:"last_name" binding:"required"`
		Active       bool   `form:"active" json:"active" xml:"active"`
		CustomerUser bool   `form:"customer_user" json:"customer_user" xml:"customer_user"`
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
	as := &service.Account{}
	data, err := as.Create(req.Username, req.Email, req.Mobile, req.FirstName, req.LastName, req.Active, req.CustomerUser)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":   http.StatusBadRequest,
			"status": http.StatusText(http.StatusBadRequest),
			"error":  err.Error(),
		})
		return
	}
	user := data["user"].(*account.User)
	profile := data["profile"].(*account.Profile)
	userProfile := &messaging.UserProfile{
		ID:            user.Model.ID,
		Username:      user.Username,
		Email:         user.Email,
		PrimaryMobile: user.Mobile,
		Active:        user.Active,
		FirstName:     profile.FirstName,
		MiddleName:    profile.MiddleName,
		LastName:      profile.LastName,
		Avatar:        user.Avatar,
	}
	if api {
		c.JSON(http.StatusCreated, gin.H{
			"code":    http.StatusCreated,
			"status":  http.StatusText(http.StatusCreated),
			"message": "User has been created",
			"data":    userProfile,
		})
		return
	}
}

func (ctrl *Account) ChangePinHandler(c *gin.Context) {
	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))
	type Request struct {
		Pin string `form:"pin" json:"pin" xml:"pin" binding:"required,min=6"`
	}
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
	as := &service.Account{}
	data, err := as.ChangeUserPin(c.GetString("user_id"), req.Pin)
	if err != nil {
		if api {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
			return
		}
	}
	if api {
		c.JSON(http.StatusCreated, gin.H{
			"code":    http.StatusAccepted,
			"status":  http.StatusText(http.StatusAccepted),
			"message": data["message"].(string),
		})
	}
}

func (ctrl *Account) DetailHandler(c *gin.Context) {
	var id string
	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))
	if api {
		id = c.Param("id")
	} else {
		id = c.Query("id")
	}

	as := &service.Account{}
	data, err := as.Get(id)
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

	user := data["user"].(*account.User)
	profile := data["profile"].(*account.Profile)
	userProfile := &messaging.UserProfile{
		ID:            user.Model.ID,
		Username:      user.Username,
		Email:         user.Email,
		PrimaryMobile: user.Mobile,
		Active:        user.Active,
		FirstName:     profile.FirstName,
		MiddleName:    profile.MiddleName,
		LastName:      profile.LastName,
		Avatar:        user.Avatar,
	}
	if api {
		c.JSON(http.StatusOK, gin.H{
			"code":   http.StatusOK,
			"status": http.StatusText(http.StatusOK),
			"data":   userProfile,
		})
	}
}

func AccountViewHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		return
	}
}

func RegisterHandler(c *gin.Context) {
	redirect := c.Query("redirect")
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, gin.H{
			"redirect": redirect,
		})
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
				"redirect": redirect,
				"error":    err.Error(),
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
	redirect := c.Query("redirect")
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, nil)
		return
	}

	as := &service.Account{}
	data, err := as.SendRegisterToken(c.GetString("userid"), "email")
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
				"redirect": redirect,
				"error":    err.Error(),
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
			"redirect": redirect,
			"referer":  data["referer"].(string),
			"purpose":  data["purpose"].(string),
			"message":  data["message"].(string),
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
	fmt.Println(c.GetString("user_id"))
	as := &service.Account{}
	data, err := as.Get(c.GetString("user_id"))
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
		} else {
			fmt.Println(err.Error())
			c.HTML(http.StatusBadRequest, constant.ProfilePage, gin.H{
				"user_id": c.GetString("user_id"),
				"error":   err.Error(),
			})
		}
		return
	}

	user := data["user"].(*account.User)
	profile := data["profile"].(*account.Profile)
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
	data, err = as.SaveProfile(c.GetString("user_id"), req.Username, req.Mobile, req.FirstName, req.LastName, req.Gender)
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
