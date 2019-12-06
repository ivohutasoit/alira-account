package service

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ivohutasoit/alira/model/domain"
)

func Login(userid, password string) (string, error) {
	if userid != "ivohutasoit" {
		return "", errors.New("invalid user or password")
	}

	if password != "hutasoit09" {
		return "", errors.New("invalid user or password")
	}

	expiresAt := time.Now().AddDate(0, 0, 7)
	userToken := &domain.UserToken{
		UserID: userid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    "Softh Axi Inc.",
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), userToken)
	tokenString, _ := token.SignedString([]byte(os.Getenv("SECRET_KEY")))

	return tokenString, nil
}
