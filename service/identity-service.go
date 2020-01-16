package service

import (
	"errors"
	"time"

	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/ivohutasoit/alira/util"
)

type IdentityService struct{}

func (s *IdentityService) CreateNationIdentity(args ...interface{}) (map[interface{}]interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("not enough parameters")
	}
	document, ok := args[0].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}

	userid, ok := args[1].(string)
	if !ok {
		return nil, errors.New("plain text parameter not type string")
	}

	user := &domain.User{}
	model.GetDatabase().First(user, "id = ? AND active = ?", userid, true)
	if user.BaseModel.ID == "" {
		return nil, errors.New("invalid user")
	}

	identity := &domain.Identity{}
	model.GetDatabase().First(identity, "user_id = ?", userid)

	nIdentity := &domain.NationalIdentity{}
	if identity.BaseModel.ID != "" {
		model.GetDatabase().First(nIdentity, "document = ? AND nation_id = ?", document, args[2].(string))
		if nIdentity.BaseModel.ID != "" {
			return nil, errors.New("identity has used other user")
		}
	} else {
		identity = &domain.Identity{
			Class:  "NATION",
			UserID: user.BaseModel.ID,
			Code:   util.GenerateNationalCode(args[4].(string), args[5].(time.Time)),
		}
		model.GetDatabase().Create(&identity)
	}

	nIdentity = &domain.NationalIdentity{
		UserID:      user.BaseModel.ID,
		IdentityID:  identity.BaseModel.ID,
		Document:    document,
		NationID:    args[2].(string),
		Fullname:    args[3].(string),
		Country:     args[4].(string),
		Nationality: "INDONESIAN",
	}
	model.GetDatabase().Create(&nIdentity)

	return map[interface{}]interface{}{
		"status":        "SUCCESS",
		"message":       "Your identity has been created",
		"identity_code": identity.Code,
		"userid":        user.BaseModel.ID,
	}, nil
}
