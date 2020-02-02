package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/ivohutasoit/alira/util"
)

type Auth struct{}

func (m *Auth) SessionRequired(args ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentPath := c.Request.URL.Path
		except := os.Getenv("WEB_EXCEPT")
		if except != "" {
			excepts := strings.Split(except, ";")
			for _, value := range excepts {
				if currentPath == strings.TrimSpace(value) {
					c.Next()
					return
				}
			}
		}

		opt := false
		optional := os.Getenv("WEB_OPTIONAL")
		if optional != "" {
			optionals := strings.Split(optional, ";")
			for _, value := range optionals {
				if value == "/" && (currentPath == "" || currentPath == "/") {
					opt = true
					break
				} else if value != "/" {
					if c.Request.Method == http.MethodGet {
						if strings.Index(currentPath, value) > -1 {
							opt = true
							return
						}
					}
				}
			}
		}

		url, err := util.GenerateUrl(c.Request.TLS, c.Request.Host, currentPath, true)
		if err != nil {
			fmt.Println(err)
		}
		redirect := fmt.Sprintf("%s?redirect=%s", os.Getenv("URL_LOGIN"), url)

		session := sessions.Default(c)
		accessToken := session.Get("access_token")
		if accessToken == nil && !opt {
			session.Clear()
			session.Save()
			c.Redirect(http.StatusMovedPermanently, redirect)
			c.Abort()
			return
		}
		if accessToken != nil {
			fmt.Println(accessToken)
			claims := &account.AccessTokenClaims{}
			token, err := jwt.ParseWithClaims(accessToken.(string), claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("SECRET_KEY")), nil
			})

			if err != nil && !opt {
				session.Clear()
				session.Save()
				c.Redirect(http.StatusMovedPermanently, redirect)
				c.Abort()
				return
			}

			if !token.Valid && !opt {
				session.Clear()
				session.Save()
				c.Redirect(http.StatusMovedPermanently, redirect)
				c.Abort()
				return
			}

			sessionToken := &account.Token{}
			alira.GetConnection().Where("access_token = ? AND valid = ?",
				accessToken, true).First(sessionToken)
			if sessionToken.Model.ID == "" && !opt {
				session.Clear()
				session.Save()
				c.Redirect(http.StatusMovedPermanently, redirect)
				c.Abort()
				return
			}

			user := &account.User{}
			alira.GetConnection().Where("id = ? AND active = ?",
				sessionToken.UserID, true).First(user)
			if user.Model.ID == "" && !opt {
				session.Clear()
				session.Save()
				c.Redirect(http.StatusMovedPermanently, redirect)
				c.Abort()
				return
			}

			c.Set("user_id", user.Model.ID)
			alira.ViewData = gin.H{
				"user_id":    user.Model.ID,
				"username":   user.Username,
				"url_logout": fmt.Sprintf("%s?redirect=%s", os.Getenv("URL_LOGOUT"), url),
			}
		}
		c.Next()
	}
}

func (m *Auth) TokenRequired(args ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentPath := c.Request.URL.Path
		urlAPI := os.Getenv("URL_API")
		except := os.Getenv("API_EXCEPT")
		if except != "" {
			excepts := strings.Split(except, ";")

			for _, value := range excepts {
				if c.Request.Method == http.MethodGet {
					if strings.Index(currentPath, value) > -1 {
						c.Next()
						return
					}
				} else if currentPath == strings.TrimSpace(fmt.Sprintf("%s%s", urlAPI, value)) {
					c.Next()
					return
				}
			}
		}

		authorization := c.Request.Header.Get("Authorization")

		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":   http.StatusUnauthorized,
				"status": http.StatusText(http.StatusUnauthorized),
				"error":  "missing authorization token",
			})
			return
		}

		tokens := strings.Split(authorization, " ")
		if len(tokens) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":   http.StatusUnauthorized,
				"status": http.StatusText(http.StatusUnauthorized),
				"error":  "invalid token",
			})
			return
		}

		var claims jwt.Claims
		if tokens[0] == "Bearer" {
			claims = &account.AccessTokenClaims{}
		} else if tokens[0] == "Refresh" {
			if currentPath != "/api/alpha/auth/refresh" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":   http.StatusUnauthorized,
					"status": http.StatusText(http.StatusUnauthorized),
					"error":  "invalid refresh uri",
				})
				return
			}
			claims = &account.RefreshTokenClaims{}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":   http.StatusUnauthorized,
				"status": http.StatusText(http.StatusUnauthorized),
				"error":  "invalid token indentifier",
			})
			return
		}

		tokenString := tokens[1]
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("SECRET_KEY")), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":   http.StatusUnauthorized,
				"status": http.StatusText(http.StatusUnauthorized),
				"error":  err.Error(),
			})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":   http.StatusUnauthorized,
				"status": http.StatusText(http.StatusUnauthorized),
				"error":  "invalid token",
			})
			return
		}
		if tokens[0] == "Refresh" {
			if claims.(*account.RefreshTokenClaims).Sub != 1 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":   http.StatusUnauthorized,
					"status": http.StatusText(http.StatusUnauthorized),
					"error":  "invalid refresh token",
				})
				return
			}
		}

		sessionToken := &account.Token{}
		if tokens[0] == "Bearer" {
			alira.GetConnection().Where("access_token = ? AND valid = ?",
				tokenString, true).First(sessionToken)
		} else if tokens[0] == "Refresh" {
			alira.GetConnection().Where("refresh_token = ? AND valid = ?",
				tokenString, true).First(sessionToken)
		}
		if sessionToken.Model.ID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":   http.StatusUnauthorized,
				"status": http.StatusText(http.StatusUnauthorized),
				"error":  "invalid token",
			})
			return
		}

		user := &account.User{}
		alira.GetConnection().Where("id = ? AND active = ?",
			sessionToken.UserID, true).First(user)
		if user.Model.ID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":   http.StatusUnauthorized,
				"status": http.StatusText(http.StatusUnauthorized),
				"error":  "invalid token",
			})
			return
		}

		c.Set("user_id", user.Model.ID)
		if tokens[0] == "Refresh" {
			c.Set("sub", claims.(*account.RefreshTokenClaims).Sub)
		}
		c.Next()
	}
}
