package service

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	alira "github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira/database/account"
)

type Token struct{}

func (s *Token) Get(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	id, ok := args[0].(string)
	if !ok {
		return nil, errors.New("parameter not type string")
	}
	token := &account.Token{}
	alira.GetConnection().First(token, "id = ? AND valid = ?",
		id, true)
	if token.Model.ID == "" {
		return nil, errors.New("invalid token")
	}
	return map[interface{}]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}, nil
}

func (s *Token) GenerateSessionToken(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 3 {
		return nil, errors.New("not enough parameters")
	}
	var userid string
	now := time.Now()
	expired := now.AddDate(0, 0, 1)
	for i, p := range args {
		switch i {
		case 1:
			param, ok := p.(time.Time)
			if !ok {
				return nil, errors.New("parameter not type time")
			}
			now = param
		case 2:
			param, ok := p.(time.Time)
			if !ok {
				return nil, errors.New("parameter not type time")
			}
			expired = param
		default:
			param, ok := p.(string)
			if !ok {
				return nil, errors.New("parameter not type string")
			}
			userid = param
		}
	}

	accessTokenClaims := &account.AccessTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        uuid.New().String(),
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiresAt: expired.Unix(),
			Issuer:    os.Getenv("ISSUER"),
		},
		UserID: userid,
		Admin:  false,
	}
	atkn := jwt.NewWithClaims(jwt.GetSigningMethod(os.Getenv("HASHING_METHOD")), accessTokenClaims)
	accessToken, _ := atkn.SignedString([]byte(os.Getenv("SECRET_KEY")))

	refreshTokenClaims := &account.RefreshTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        uuid.New().String(),
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiresAt: expired.AddDate(0, 0, 1).Unix(),
			Issuer:    os.Getenv("ISSUER"),
		},
		UserID: userid,
		Sub:    1,
	}
	rtkn := jwt.NewWithClaims(jwt.GetSigningMethod(os.Getenv("HASHING_METHOD")), refreshTokenClaims)
	refreshToken, _ := rtkn.SignedString([]byte(os.Getenv("SECRET_KEY")))

	return map[interface{}]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

func (s *Token) GetAuthenticated(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameter")
	}
	tokenType, ok := args[0].(string)
	if !ok {
		return nil, errors.New("parameter not type string")
	}
	tokenString, ok := args[1].(string)
	if !ok {
		return nil, errors.New("parameter not type string")
	}

	token := &account.Token{}
	user := &account.User{}
	if tokenType == "Bearer" {
		alira.GetConnection().First(token, "access_token = ? AND valid = ?",
			tokenString, true)
	} else {
		alira.GetConnection().First(&token, "refresh_token = ? AND valid = ?",
			tokenString, true)
	}
	if token.Model.ID == "" {
		return nil, errors.New("invalid token")
	}

	alira.GetConnection().First(user, "id = ? AND active = ?",
		token.UserID, true)
	if user.Model.ID == "" {
		return nil, errors.New("invalid token")
	}

	return map[interface{}]interface{}{
		"status": "success",
		"valid":  true,
		"user":   user,
	}, nil
}
