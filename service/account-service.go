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

type AccountService struct{}

func (as *AccountService) SaveRegisterToken(args ...interface{}) (*domain.Token, error) {
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
	if sentTo == "email" {
		mail := &domain.Mail{
			From:     os.Getenv("SMTP.SENDER"),
			To:       []string{token.Referer},
			Subject:  "Token Registration",
			Template: "templates/mail/token_registration.html",
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

func (ac *AccountService) ActivateToken(args ...interface{}) (map[string]string, error) {
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
	if token == nil || token.NotAfter.Before(time.Now()) {
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
		Purpose:   "BASIC USAGE",
		Signature: util.GenerateToken(16),
		NotBefore: time.Now(),
		AgreedAt:  time.Now(),
	}

	sessionToken := &domain.Token{
		BaseModel: model.BaseModel{
			ID: uuid.New().String(),
		},
		Class:     "SESSION",
		UserID:    user.ID,
		User:      *user,
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 12),
		Valid:     true,
	}
	authService := &AuthService{}
	tokens, err := authService.Login("Creation", token)
	if err != nil {
		return nil, err
	}
	sessionToken.AccessToken = tokens["access_token"].(string)
	sessionToken.RefreshToken = tokens["refresh_token"].(string)

	model.GetDatabase().Delete(token)
	model.GetDatabase().Create(user)
	model.GetDatabase().Create(profile)
	model.GetDatabase().Create(subscribe)
	model.GetDatabase().Create(sessionToken)

	return map[string]string{
		"user":          user.ID,
		"access_token":  sessionToken.AccessToken,
		"refresh_token": sessionToken.RefreshToken,
	}, nil
}
