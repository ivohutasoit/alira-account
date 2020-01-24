package service

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
)

type TokenService struct{}

func (s *TokenService) GenerateSessionToken(args ...interface{}) (map[interface{}]interface{}, error) {
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

	accessTokenClaims := &domain.AccessTokenClaims{
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

	refreshTokenClaims := &domain.RefreshTokenClaims{
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

func (s *TokenService) GetTokenInformation(args ...interface{}) (map[interface{}]interface{}, error) {
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

	token := &domain.Token{}
	user := &domain.User{}
	if tokenType == "Bearer" {
		model.GetDatabase().First(token, "access_token = ? AND valid = ?",
			tokenString, true)
	} else {
		model.GetDatabase().First(&token, "refresh_token = ? AND valid = ?",
			tokenString, true)
	}
	if token.BaseModel.ID == "" {
		return nil, errors.New("invalid token")
	}

	model.GetDatabase().First(user, "id = ? AND active = ? AND deleted_at IS NULL",
		token.UserID, true)
	if user.BaseModel.ID == "" {
		return nil, errors.New("invalid token")
	}

	return map[interface{}]interface{}{
		"status": "success",
		"valid":  true,
		"user":   user,
	}, nil
}
