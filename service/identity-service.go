package service

import (
	"errors"

	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
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

	nIdentity := &domain.NationalIdentity{}
	model.GetDatabase().First(nIdentity, "document = ?", document)
	if nIdentity.BaseModel.ID != "" {
		return nil, errors.New("identity has used other user")
	}

	identity := &domain.Identity{}
	model.GetDatabase().First(identity, "user_id = ?", userid)
	if identity.BaseModel.ID != "" {
		return nil, errors.New("user has been identified")
	}

	identity = &domain.Identity{
		Class:  "NATION",
		UserID: user.BaseModel.ID,
	}
	nIdentity = &domain.NationalIdentity{
		UserID: user.BaseModel.ID,
	}

	return nil, nil
}
