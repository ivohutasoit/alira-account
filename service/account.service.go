package service

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/ivohutasoit/alira/service"
	"github.com/ivohutasoit/alira/util"
)

type Account struct{}

func (s *Account) Create(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	var username, email, mobile, firstName, lastName string
	active, customerUser := false, false
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
		case 5:
			param, ok := v.(bool)
			if !ok {
				return nil, errors.New("plain parameter not type bool")
			}
			active = param
		case 6:
			param, ok := v.(bool)
			if !ok {
				return nil, errors.New("plain parameter not type bool")
			}
			customerUser = param
		default:
			param, ok := v.(string)
			if !ok {
				return nil, errors.New("plain text parameter not type string")
			}
			username = strings.ToLower(strings.TrimSpace(param))
		}
	}
	var users []account.User
	alira.GetConnection().Where("username = ? OR email = ? OR mobile = ?",
		username, email, mobile).Find(&users)
	if len(users) > 0 {
		return nil, errors.New("username has been taken")
	}
	user := &account.User{
		Username:       username,
		Email:          email,
		Mobile:         mobile,
		Active:         active,
		FirstTimeLogin: customerUser,
	}
	alira.GetConnection().Create(user)

	profile := &account.Profile{
		ID:        user.Model.ID,
		FirstName: firstName,
		LastName:  lastName,
	}
	alira.GetConnection().Create(profile)

	return map[interface{}]interface{}{
		"status":  "SUCCESS",
		"user":    user,
		"profile": profile,
	}, nil
}

func (s *Account) Get(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	userid, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}
	user := &account.User{}
	alira.GetConnection().Where("id = ?", userid).First(&user)
	profile := &account.Profile{}
	alira.GetConnection().Where("id = ?", userid).First(&profile)

	if user.Model.ID == "" {
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

func (s *Account) ChangeUserPin(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameter")
	}
	userid, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text is not type string")
	}
	pin, ok := args[1].(string)
	if !ok {
		return nil, errors.New("plain text is not type string")
	}
	user := &account.User{}
	alira.GetConnection().Where("id = ? AND active = ?",
		userid, true).First(&user)
	if user.Model.ID == "" {
		return nil, errors.New("invalid user")
	}
	if !user.UsePin {
		user.UsePin = true
	}
	user.Pin = strings.TrimSpace(pin)
	alira.GetConnection().Save(&user)

	return map[interface{}]interface{}{
		"message": "User pin has been changed",
		"user":    user,
	}, nil
}

func (as *Account) SendRegisterToken(args ...interface{}) (map[interface{}]interface{}, error) {
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

	user := &account.User{}
	alira.GetConnection().First(user, "active = ? AND (username = ? OR email = ? OR mobile = ?)",
		true, payload, payload, payload)
	if user.Model.ID != "" {
		return nil, errors.New("user already exists")
	}

	token := &account.Token{}
	alira.GetConnection().First(token, "valid = ? AND class = ? AND referer = ?", true, "REGISTER", payload)

	if token != nil {
		token.Valid = false
		alira.GetConnection().Save(&token)
	}

	token = &account.Token{
		Referer:     payload,
		Class:       "REGISTER",
		AccessToken: util.GenerateToken(6),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour * 12),
		Valid:       true,
	}
	if sentTo == "email" {
		mail := &service.Mail{
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

	alira.GetConnection().Create(token)

	return map[interface{}]interface{}{
		"status":  "SUCCESS",
		"purpose": token.Class,
		"referer": token.Referer,
		"message": "Registration token has been sent to your email",
	}, nil
}

func (ac *Account) ActivateRegistration(args ...interface{}) (map[interface{}]interface{}, error) {
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
	token := &account.Token{}
	alira.GetConnection().First(token, "access_token = ? AND referer = ? AND valid = ? AND class = ?",
		code, referer, true, "REGISTER")
	if token == nil {
		return nil, errors.New("invalid token")
	}

	user := &account.User{
		Email:  token.Referer,
		Active: true,
	}

	profile := &account.Profile{}

	subscribe := &account.Subscription{
		Code:      "BASIC",
		Name:      "Basic Account",
		Signature: util.GenerateToken(16),
		NotBefore: time.Now(),
		AgreedAt:  time.Now(),
	}

	sessionToken := &account.Token{
		Class:     "SESSION",
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 12),
		Valid:     true,
	}
	now := time.Now()
	expired := now.AddDate(0, 0, 1)

	ts := &TokenService{}
	data, err := ts.GenerateSessionToken(user.Model.ID, now, expired)
	if err != nil {
		return nil, err
	}

	sessionToken.AccessToken = data["access_token"].(string)
	sessionToken.RefreshToken = data["refresh_token"].(string)

	alira.GetConnection().Create(&user)

	profile.ID = user.Model.ID
	alira.GetConnection().Create(&profile)

	subscribe.Subscriber = user.Model.ID
	alira.GetConnection().Create(&subscribe)

	sessionToken.UserID = user.Model.ID
	alira.GetConnection().Create(&sessionToken)

	token.UserID = user.Model.ID
	token.Valid = false
	alira.GetConnection().Save(&token)

	return map[interface{}]interface{}{
		"user_id":       user.Model.ID,
		"email":         user.Email,
		"access_token":  sessionToken.AccessToken,
		"refresh_token": sessionToken.RefreshToken,
	}, nil
}

func (s *Account) SaveProfile(args ...interface{}) (map[interface{}]interface{}, error) {
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
	user := &account.User{}
	alira.GetConnection().First(user, "id = ? AND active = ?", userid, true)

	if user.Model.ID == "" {
		return nil, errors.New("invalid user")
	}

	profile := &account.Profile{}
	alira.GetConnection().First(profile, "id = ?", user.Model.ID)
	if profile == nil {
		return nil, errors.New("invalid user profile")
	}

	user.Username = strings.TrimSpace(username)
	user.Mobile = strings.TrimSpace(mobile)
	alira.GetConnection().Save(&user)

	profile.FirstName = strings.TrimSpace(firstName)
	profile.LastName = strings.TrimSpace(lastName)
	profile.Gender = strings.TrimSpace(gender)
	alira.GetConnection().Save(&profile)

	return map[interface{}]interface{}{
		"user_id": userid,
		"message": "User profile has been saved succesfully",
	}, nil
}
