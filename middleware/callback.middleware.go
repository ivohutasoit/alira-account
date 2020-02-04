package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/ivohutasoit/alira/util"
)

type Callback struct{}

func (m *Callback) ValidateSession(args ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		accessToken := session.Get("access_token")
		if accessToken != nil {
			token := &account.Token{}
			alira.GetConnection().Where("access_token = ? AND valid = ?",
				accessToken, true).First(&token)
			if token.Model.ID != "" {
				redirect := c.Query("redirect")
				if redirect != "" {
					uri, _ := util.Decrypt(redirect, os.Getenv("SECRET_KEY"))
					if strings.Contains(uri, "?") {
						uri = fmt.Sprintf("%s&callback=%s", uri, token.Model.ID)
					} else {
						uri = fmt.Sprintf("%s?callback=%s", uri, token.Model.ID)
					}
					c.Redirect(http.StatusMovedPermanently, uri)
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}
