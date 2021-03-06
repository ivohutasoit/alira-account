package service

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira-account/model"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/ivohutasoit/alira/service"
	"github.com/ivohutasoit/alira/util"
)

type Auth struct{}

func (s *Auth) AuthenticateUser(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	param, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}
	user := &account.User{}
	alira.GetConnection().Where("id = ? AND (active = ? OR first_time_login = ?)",
		param, true, true).First(&user)
	if user.Model.ID == "" {
		userid := strings.ToLower(param)
		alira.GetConnection().Debug().Where("(username = ? OR email = ? OR mobile = ?) AND active = ?",
			userid, userid, userid, true).First(&user)
	}
	if user.Model.ID == "" {
		return nil, errors.New("invalid login")
	}

	if !user.UsePin {
		token := &account.Token{}
		alira.GetConnection().Where("valid = ? AND class = ? AND user_id = ?",
			true, "LOGIN", user.ID).First(&token)
		if token.Model.ID != "" {
			token.Valid = false
			alira.GetConnection().Save(&token)
		}

		loginToken := &account.Token{
			Class:       "LOGIN",
			Referer:     user.Model.ID,
			UserID:      user.Model.ID,
			AccessToken: util.GenerateToken(6),
			NotBefore:   time.Now(),
			NotAfter:    time.Now().Add(time.Minute * 5),
			Valid:       true,
		}
		alira.GetConnection().Create(loginToken)
		mail := &service.Mail{
			From:     os.Getenv("SMTP_SENDER"),
			To:       []string{user.Email},
			Subject:  "[Alira] Authentication Token",
			Template: "views/mail/login.html",
			Data: map[interface{}]interface{}{
				"username": user.Username,
				"token":    loginToken.AccessToken,
				"interval": "5 minutes",
			},
		}
		ms := &service.MailService{}
		_, err := ms.Send(mail)
		if err != nil {
			return nil, err
		}
	}
	return map[interface{}]interface{}{
		"purpose": "LOGIN",
		"user":    user,
	}, nil
}

func (s *Auth) GenerateLoginSocket(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	param, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}

	code := util.GenerateQrcode(16)

	encrypted, err := util.Encrypt(code, os.Getenv("SECRET_KEY"))
	if err != nil {
		return nil, err
	}
	model.Sockets[code] = model.LoginSocket{
		Redirect: param,
		Status:   1,
	}

	return map[interface{}]interface{}{
		"code": encrypted,
	}, nil
}

func (s *Auth) SendLoginToken(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	param, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}
	userid := strings.ToLower(param)
	user := &account.User{}
	alira.GetConnection().First(user, "(username = ? OR email = ? OR mobile = ?) and active = ?",
		userid, userid, userid, true)
	if user.Model.ID == "" {
		return nil, errors.New("invalid user")
	}

	token := &account.Token{}
	alira.GetConnection().First(token, "valid = ? AND class = ? AND user_id = ?",
		true, "LOGIN", user.ID)
	if token.Model.ID != "" {
		token.Valid = false
		alira.GetConnection().Save(&token)
	}

	token = &account.Token{
		Class:       "LOGIN",
		Referer:     user.Model.ID,
		UserID:      user.Model.ID,
		AccessToken: util.GenerateToken(6),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Minute * 5),
		Valid:       true,
	}
	alira.GetConnection().Create(token)
	mail := &service.Mail{
		From:     os.Getenv("SMTP_SENDER"),
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
		"status":  "SUCCESS",
		"purpose": "LOGIN",
		"referer": user.Model.ID,
		"message": "Token login has been sent to your email",
	}, nil
}

func (s *Auth) VerifyToken(args ...interface{}) (map[interface{}]interface{}, error) {
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
	var user account.User
	alira.GetConnection().Where("id = ? AND (active = ? OR first_time_login = ?)",
		userid, true, true).First(&user)
	if user.Model.ID == "" {
		return nil, errors.New("invalid login")
	}

	if user.UsePin && user.Pin != code {
		return nil, errors.New("invalid login")
	} else if !user.UsePin {
		var token account.Token
		if user.FirstTimeLogin {
			alira.GetConnection().Where("access_token = ? AND user_id = ? AND valid = ?",
				code, userid, true).First(&token)
			if token.Model.ID == "" {
				return nil, errors.New("invalid token")
			}
			token.Valid = false
			alira.GetConnection().Save(&token)
		}
	}

	var oldSessionToken account.Token
	alira.GetConnection().Where("user_id = ? AND valid = ?",
		user.Model.ID, true).First(&oldSessionToken)
	if oldSessionToken.Model.ID != "" {
		oldSessionToken.Valid = false
		alira.GetConnection().Save(&oldSessionToken)
	}

	now := time.Now()
	expired := now.AddDate(0, 0, 1)

	ts := &TokenService{}
	data, err := ts.GenerateSessionToken(user.Model.ID, now, expired)
	if err != nil {
		return nil, err
	}

	sessionToken := &account.Token{
		Class:        "SESSION",
		UserID:       user.Model.ID,
		AccessToken:  data["access_token"].(string),
		RefreshToken: data["refresh_token"].(string),
		NotBefore:    now,
		NotAfter:     expired,
		Valid:        true,
	}
	alira.GetConnection().Create(&sessionToken)

	if user.FirstTimeLogin {
		user.Active = true
		user.FirstTimeLogin = false
		alira.GetConnection().Save(&user)
	}

	data["user"] = user
	if user.Username == "" {
		data["profile"] = "required"
	} else {
		data["profile"] = "completed"
	}

	return data, nil
}

func (s *Auth) GenerateRefreshToken(args ...interface{}) (map[interface{}]interface{}, error) {
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
	token := &account.Token{}
	alira.GetConnection().First(token, "refresh_token = ? AND user_id = ? AND valid = ?",
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

	sessionToken := &account.Token{
		Class:        "SESSION",
		UserID:       userid,
		AccessToken:  data["access_token"].(string),
		RefreshToken: data["refresh_token"].(string),
		NotBefore:    now,
		NotAfter:     expired,
		Valid:        true,
	}
	alira.GetConnection().Create(sessionToken)

	token.Valid = false
	alira.GetConnection().Update(token)

	return data, nil
}

func (s *Auth) RemoveSessionToken(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}

	accessToken := args[0].(string)
	if accessToken == " " {
		return nil, errors.New("invalid token")
	}

	token := &account.Token{}
	alira.GetConnection().First(token, "access_token = ? AND valid = ?", accessToken, true)
	if token == nil {
		return nil, errors.New("invalid token")
	}

	token.Valid = false
	alira.GetConnection().Save(&token)

	return map[interface{}]interface{}{
		"status":  "success",
		"message": "log out successful and please log in to get full access to your account",
	}, nil
}
