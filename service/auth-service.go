package service

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ivohutasoit/alira/model/domain"
)

type AuthService struct {
}

func (s *AuthService) Login(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	authType, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}

	var userid string
	if authType == "Basic" {
		if len(args) < 3 {
			return nil, errors.New("not enough parameters")
		}

		userid, ok := args[1].(string)
		if !ok {
			return nil, errors.New("plain text parameter not type string")
		}
		if userid != "ivohutasoit" {
			return nil, errors.New("invalid user or password")
		}

		password, ok := args[2].(string)
		if !ok {
			return nil, errors.New("plain text parameter not type string")
		}
		if password != "hutasoit09" {
			return nil, errors.New("invalid user or password")
		}

	} else if authType == "Refresh" {
		if len(args) < 2 {
			return nil, errors.New("not enough parameters")
		}

		userid, ok := args[1].(string)
		if !ok {
			return nil, errors.New("plain text parameter not type string")
		}
		if userid != "ivohutasoit" {
			return nil, errors.New("invalid user or password")
		}
	}

	accessTokenClaims := &domain.AccessTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        userid,
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().AddDate(0, 0, 7).Unix(),
			Issuer:    os.Getenv("ISSUER"),
		},
		Admin: false,
	}
	atkn := jwt.NewWithClaims(jwt.GetSigningMethod(os.Getenv("HASHING_METHOD")), accessTokenClaims)
	accessToken, _ := atkn.SignedString([]byte(os.Getenv("SECRET_KEY")))

	refreshTokenClaims := &domain.RefreshTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        userid,
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().AddDate(0, 0, 8).Unix(),
			Issuer:    os.Getenv("ISSUER"),
		},
		Sub: 1,
	}
	rtkn := jwt.NewWithClaims(jwt.GetSigningMethod(os.Getenv("HASHING_METHOD")), refreshTokenClaims)
	refreshToken, _ := rtkn.SignedString([]byte(os.Getenv("SECRET_KEY")))

	return map[interface{}]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}
