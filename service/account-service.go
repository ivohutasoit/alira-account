package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/ivohutasoit/alira/service"
	"github.com/ivohutasoit/alira/util"
)

type AccountService struct{}

func (s *AccountService) Create(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	var username, email, mobile, firstName, lastName string
	for i, v := range args {
		switch i {
		case 1:
			param, ok := v.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			email = strings.ToLower(strings.TrimSpace(param))
		case 2:
			param, ok := v.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			mobile = strings.TrimSpace(param)
		case 3:
			param, ok := v.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			firstName = strings.Title(strings.TrimSpace(param))
		case 4:
			param, ok := v.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			lastName = strings.Title(strings.TrimSpace(param))
		default:
			param, ok := v.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			username = strings.ToLower(strings.TrimSpace(param))
		}
	}
	fmt.Println("Step 1")
	var users []domain.User
	model.GetDatabase().Where("active = ? AND (username = ? OR email = ? OR mobile = ?)",
		true, username, email, mobile).Find(&users)
	if len(users) > 0 {
		return nil, errors.New("username has been taken")
	}
	fmt.Println("Step 2")
	user := &domain.User{
		Username: username,
		Email:    email,
		Mobile:   mobile,
		Active:   false,
	}
	model.GetDatabase().Create(user)

	fmt.Println("Step 3")
	profile := &domain.Profile{
		ID:        user.BaseModel.ID,
		FirstName: firstName,
		LastName:  lastName,
	}
	model.GetDatabase().Create(profile)

	fmt.Println("Step 4")
	return map[interface{}]interface{}{
		"status": "SUCCESS",
		"user":   user,
	}, nil
}

func (s *AccountService) Get(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	userid, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}
	user := &domain.User{}
	model.GetDatabase().First(&user, "id = ?", userid)
	profile := &domain.Profile{}
	model.GetDatabase().First(&profile, "id = ?", userid)

	if user.BaseModel.ID == "" {
		return nil, errors.New("invalid user")
	}

	if profile.ID == "" {
		return nil, errors.New("invalid user profile")
	}

	return map[interface{}]interface{}{
		"user":    user,
		"profile": profile,
	}, nil
}

func (as *AccountService) SendRegisterToken(args ...interface{}) (map[interface{}]interface{}, error) {
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
	if user.BaseModel.ID != "" {
		return nil, errors.New("user already exists")
	}

	token := &domain.Token{}
	model.GetDatabase().First(token, "valid = ? AND class = ? AND referer = ?", true, "REGISTER", payload)

	if token != nil {
		token.Valid = false
		model.GetDatabase().Save(&token)
	}

	token = &domain.Token{
		Referer:     payload,
		Class:       "REGISTER",
		AccessToken: util.GenerateToken(6),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour * 12),
		Valid:       true,
	}
	if sentTo == "email" {
		mail := &domain.Mail{
			From:     os.Getenv("SMTP_SENDER"),
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

	return map[interface{}]interface{}{
		"status":  "SUCCESS",
		"purpose": token.Class,
		"referer": token.Referer,
		"message": "Registration token has been sent to your email",
	}, nil
}

func (ac *AccountService) ActivateRegistration(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameters")
	}
	var referer, code string
	for i, p := range args {
		param, ok := p.(string)
		if !ok {
			return nil, errors.New("plain text parameter not type string")
		}
		switch i {
		case 1:
			code = param
			break
		default:
			referer = param
			break
		}
	}
	token := &domain.Token{}
	model.GetDatabase().First(token, "access_token = ? AND referer = ? AND valid = ? AND class = ?",
		code, referer, true, "REGISTER")
	if token == nil {
		return nil, errors.New("invalid token")
	}

	user := &domain.User{
		Email:  token.Referer,
		Active: true,
	}

	profile := &domain.Profile{}

	subscribe := &domain.Subscribe{
		Code:      "BASIC",
		Name:      "Basic Account",
		Signature: util.GenerateToken(16),
		NotBefore: time.Now(),
		AgreedAt:  time.Now(),
	}

	sessionToken := &domain.Token{
		Class:     "SESSION",
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

	model.GetDatabase().Create(&user)

	profile.ID = user.BaseModel.ID
	model.GetDatabase().Create(&profile)

	subscribe.SubscriberID = user.BaseModel.ID
	model.GetDatabase().Create(&subscribe)

	sessionToken.UserID = user.BaseModel.ID
	model.GetDatabase().Create(&sessionToken)

	token.UserID = user.BaseModel.ID
	token.Valid = false
	model.GetDatabase().Save(&token)

	return map[interface{}]interface{}{
		"user_id":       user.BaseModel.ID,
		"email":         user.Email,
		"access_token":  sessionToken.AccessToken,
		"refresh_token": sessionToken.RefreshToken,
	}, nil
}

func (s *AccountService) SaveProfile(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 6 {
		return nil, errors.New("not enough parameters")
	}

	var userid, username, mobile, firstName, lastName, gender string
	for i, p := range args {
		param, ok := p.(string)
		if !ok {
			return nil, errors.New("plain text parameter not type string")
		}
		switch i {
		case 1:
			username = strings.ToLower(param)
			break
		case 2:
			mobile = param
			break
		case 3:
			firstName = strings.Title(param)
			break
		case 4:
			lastName = strings.Title(param)
			break
		case 5:
			gender = strings.ToLower(param)
			break
		default:
			userid = param
		}
	}
	user := &domain.User{}
	model.GetDatabase().First(user, "id = ? AND active = ?", userid, true)

	if user.BaseModel.ID == "" {
		return nil, errors.New("invalid user")
	}

	profile := &domain.Profile{}
	model.GetDatabase().First(profile, "id = ?", user.BaseModel.ID)
	if profile == nil {
		return nil, errors.New("invalid user profile")
	}

	user.Username = strings.TrimSpace(username)
	user.Mobile = strings.TrimSpace(mobile)
	model.GetDatabase().Save(&user)

	profile.Name = strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
	profile.FirstName = strings.TrimSpace(firstName)
	profile.LastName = strings.TrimSpace(lastName)
	profile.Gender = strings.TrimSpace(gender)
	model.GetDatabase().Save(&profile)

	return map[interface{}]interface{}{
		"user_id": userid,
		"message": "User profile has been saved succesfully",
	}, nil
}
