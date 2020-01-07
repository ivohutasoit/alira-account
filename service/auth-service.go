package service

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/ivohutasoit/alira/service"
	"github.com/ivohutasoit/alira/util"
)

type AuthService struct{}

func (s *AuthService) SendLoginToken(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	userid, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}
	user := &domain.User{}
	model.GetDatabase().First(user, "(username = ? OR email = ? OR mobile = ?) and active = ?",
		userid, userid, userid, true)

	if user == nil {
		return nil, errors.New("invalid user or password")
	}
	token := &domain.Token{
		BaseModel: model.BaseModel{
			ID: uuid.New().String(),
		},
		Class:       "LOGIN",
		Referer:     user.BaseModel.ID,
		UserID:      user.BaseModel.ID,
		User:        *user,
		AccessToken: util.GenerateToken(6),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Minute * 5),
		Valid:       true,
	}
	model.GetDatabase().Create(token)
	mail := &domain.Mail{
		From:     os.Getenv("SMTP.SENDER"),
		To:       []string{user.Email},
		Subject:  "Login Token",
		Template: "views/mail/token_login.html",
		Data: map[interface{}]interface{}{
			"username": user.Email,
			"token":    token.AccessToken,
			"interval": "5 minutes",
		},
	}
	ms := &service.MailService{}
	_, err := ms.Send(mail)
	if err != nil {
		return nil, err
	}
	return map[interface{}]interface{}{
		"status":  "success",
		"purpose": "LOGIN",
		"message": "Token login has been sent to your email",
	}, nil
}

func (s *AuthService) VerifyLoginToken(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameters")
	}
	var userid, code string
	for i, p := range args {
		switch i {
		case 1:
			param, ok := p.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			code = param
			break
		default:
			param, ok := p.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			userid = param
			break
		}
	}
	token := &domain.Token{}
	model.GetDatabase().First(token, "access_token = ? AND user_id = ? AND valid = ?",
		code, userid, true)
	if token == nil {
		return nil, errors.New("invalid token")
	}

	now := time.Now()
	expired := now.AddDate(0, 0, 1)

	ts := &TokenService{}
	data, err := ts.GenerateToken(userid, now, expired)
	if err != nil {
		return nil, err
	}

	sessionToken := &domain.Token{
		BaseModel: model.BaseModel{
			ID: uuid.New().String(),
		},
		Class:        "SESSION",
		UserID:       userid,
		AccessToken:  data["access_token"].(string),
		RefreshToken: data["refresh_token"].(string),
		NotBefore:    now,
		NotAfter:     expired,
		Valid:        true,
	}
	model.GetDatabase().Create(sessionToken)

	token.Valid = false
	model.GetDatabase().Update(token)

	return data, nil
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
	now := time.Now()
	expired := now.AddDate(0, 0, 1)
	user := &domain.User{}
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

		/*password, ok := args[2].(string)
		if !ok {
			return nil, errors.New("plain text parameter not type string")
		}
		if password != "hutasoit09" {
			return nil, errors.New("invalid user or password")
		}*/

		model.GetDatabase().First(user, "(username = ? OR email = ? OR mobile = ?) and active = ?",
			userid, userid, userid, true)

		if user == nil {
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
	} else if authType == "Creation" {
		if len(args) < 2 {
			return nil, errors.New("not enough parameters")
		}
		param, ok := args[1].(*domain.Token)
		if !ok {
			return nil, errors.New("parameter not type token")
		}
		userid = param.UserID
		now = param.NotBefore
		expired = param.NotAfter
	}

	accessTokenClaims := &domain.AccessTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        user.BaseModel.ID,
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiresAt: expired.Unix(),
			Issuer:    os.Getenv("ISSUER"),
		},
		Admin: false,
	}
	atkn := jwt.NewWithClaims(jwt.GetSigningMethod(os.Getenv("HASHING_METHOD")), accessTokenClaims)
	accessToken, _ := atkn.SignedString([]byte(os.Getenv("SECRET_KEY")))

	refreshTokenClaims := &domain.RefreshTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        userid,
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiresAt: expired.AddDate(0, 0, 1).Unix(),
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
