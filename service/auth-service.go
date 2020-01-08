package service

import (
	"errors"
	"os"
	"time"

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
		Subject:  "[Alira] Authentication Token",
		Template: "views/mail/login.html",
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
	data, err := ts.GenerateSessionToken(userid, now, expired)
	if err != nil {
		return nil, err
	}

	sessionToken := &domain.Token{
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

func (s *AuthService) GenerateRefreshToken(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameters")
	}
	var userid, refreshToken string
	for i, p := range args {
		switch i {
		case 1:
			param, ok := p.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			refreshToken = param
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
	model.GetDatabase().First(token, "refresh_token = ? AND user_id = ? AND valid = ?",
		refreshToken, userid, true)
	if token == nil {
		return nil, errors.New("invalid refresh token")
	}
	now := time.Now()
	expired := now.AddDate(0, 0, 1)

	ts := &TokenService{}
	data, err := ts.GenerateSessionToken(userid, now, expired)
	if err != nil {
		return nil, err
	}

	sessionToken := &domain.Token{
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
