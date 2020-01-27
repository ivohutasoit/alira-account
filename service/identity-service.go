package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

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

	identity := &domain.Identity{}
	model.GetDatabase().First(identity, "user_id = ?", userid)

	nIdentity := &domain.NationalIdentity{}
	if identity.BaseModel.ID != "" {
		model.GetDatabase().First(nIdentity, "document = ? AND nation_id = ?", document, args[2].(string))
		if nIdentity.BaseModel.ID != "" {
			return nil, errors.New("identity has used other user")
		}
	} else {
		code, _ := GenerateNationalCode(args[4].(string), args[5].(time.Time))
		identity = &domain.Identity{
			Class:  "NATION",
			UserID: user.BaseModel.ID,
			Code:   code.(string),
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

func GenerateNationalCode(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, errors.New("not enough parameters")
	}
	//Example Nomor NIK : 1234567890ABCDEF
	//12 nomor merupakan kode provinsi
	//34 nomor merupakan kode kotamadya atau kabupaten kota
	//56 nomor kode kecamatan
	//78 nomor tanggal lahir
	//90 nomor bulan lahir
	//AB nomor tahun lahir
	//CDEF nomor registrasi kependudukan

	// Country code 2 characters ID=62
	// BirthDate 8 characters format yyyyMMDD
	// Sequence 4 characters
	date, _ := args[1].(time.Time)
	var country int
	if args[0].(string) == "INDONESIA" {
		country = 62
	}

	code := fmt.Sprintf("%d%d%02d%02d", country, date.Year(), date.Month(), date.Day())

	var identities []domain.Identity
	model.GetDatabase().Where("code LIKE ?", code+"%").Find(&identities).Order("code DESC")

	var identity domain.Identity
	if identities != nil && len(identities) < 0 {
		identity = identities[0]
	}

	var seq int
	if identity.BaseModel.ID != "" {
		if n, err := strconv.Atoi(identity.Code[10:len(identity.Code)]); err == nil {
			seq = n + 1
		} else {
			return nil, errors.New("is not an integer")
		}
	} else {
		seq = 1
	}

	return fmt.Sprintf("%s%06d", code, seq), nil
}
