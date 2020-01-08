package service

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/ivohutasoit/alira/service"
	"github.com/ivohutasoit/alira/util"
)

type AccountService struct{}

func (as *AccountService) SendRegisterToken(args ...interface{}) (*domain.Token, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameters")
	}

	var payload, sentTo string
	for i, p := range args {
		switch i {
		case 1:
			param, ok := p.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			sentTo = param
			break
		default:
			param, ok := p.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type token")
			}
			payload = param
			break
		}
	}

	user := &domain.User{}
	model.GetDatabase().First(user, "active = ? AND (username = ? OR email = ? OR mobile = ?)",
		true, payload, payload, payload)
	if user.ID != "" {
		return nil, errors.New("user already exists")
	}

	token := &domain.Token{
		BaseModel: model.BaseModel{
			ID: uuid.New().String(),
		},
		Referer:     payload,
		Class:       "REGISTER",
		AccessToken: util.GenerateToken(6),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour * 12),
		Valid:       true,
	}
	fmt.Println(token.Referer)
	if sentTo == "email" {
		mail := &domain.Mail{
			From:     os.Getenv("SMTP.SENDER"),
			To:       []string{token.Referer},
			Subject:  "[Alira] Registration Token",
			Template: "views/mail/registration.html",
			Data: map[interface{}]interface{}{
				"token": token.AccessToken,
			},
		}
		ms := &service.MailService{}
		_, err := ms.Send(mail)
		if err != nil {
			return nil, err
		}
	}

	model.GetDatabase().Create(token)

	return token, nil
}

func (ac *AccountService) ActivateRegistration(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameters")
	}
	var referer, code string
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
			referer = param
			break
		}
	}
	token := &domain.Token{}
	model.GetDatabase().First(token, "access_token = ? AND referer = ? AND valid = ?",
		code, referer, true)
	if token == nil {
		return nil, errors.New("invalid token")
	}

	user := &domain.User{
		BaseModel: model.BaseModel{
			ID: uuid.New().String(),
		},
		Email:  token.Referer,
		Active: true,
	}

	profile := &domain.Profile{
		BaseModel: model.BaseModel{
			ID: user.BaseModel.ID,
		},
		User: *user,
	}

	subscribe := &domain.Subscribe{
		BaseModel: model.BaseModel{
			ID: uuid.New().String(),
		},
		Code:      "BASIC",
		UserID:    user.BaseModel.ID,
		User:      *user,
		Purpose:   "Basic Account Usage",
		Signature: util.GenerateToken(16),
		NotBefore: time.Now(),
		AgreedAt:  time.Now(),
	}

	sessionToken := &domain.Token{
		BaseModel: model.BaseModel{
			ID: uuid.New().String(),
		},
		Class:     "SESSION",
		UserID:    user.BaseModel.ID,
		User:      *user,
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 12),
		Valid:     true,
	}
	now := time.Now()
	expired := now.AddDate(0, 0, 1)

	ts := &TokenService{}
	data, err := ts.GenerateSessionToken(user.BaseModel.ID, now, expired)
	if err != nil {
		return nil, err
	}

	sessionToken.AccessToken = data["access_token"].(string)
	sessionToken.RefreshToken = data["refresh_token"].(string)

	token.UserID = user.BaseModel.ID
	token.User = *user
	token.Valid = false

	model.GetDatabase().Create(user)
	model.GetDatabase().Create(profile)
	model.GetDatabase().Create(subscribe)
	model.GetDatabase().Create(sessionToken)
	model.GetDatabase().Update(token)

	return map[interface{}]interface{}{
		"user":          user.BaseModel.ID,
		"access_token":  sessionToken.AccessToken,
		"refresh_token": sessionToken.RefreshToken,
	}, nil
}
