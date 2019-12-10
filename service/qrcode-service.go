package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/ivohutasoit/alira/common"
)

type QrcodeService struct {
}

func (s *QrcodeService) Generate() string {
	b := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = common.Letters[rand.Intn(len(common.Letters))]
	}
	return string(b)
}

func (s *QrcodeService) Verify(args ...interface{}) (string, error) {
	if 1 > len(args) {
		return "", errors.New("not enough parameters")
	}

	var code, token string
	for i, p := range args {
		switch i {
		case 0:
			param, ok := p.(string)
			if !ok {
				return "", errors.New("string type required")
			}
			code = param
		case 1:
			param, ok := p.(string)
			if !ok {
				return "", errors.New("string type required")
			}
			token = param
		}

	}
	fmt.Printf("%s %s", code, token)
	return "", nil
}
