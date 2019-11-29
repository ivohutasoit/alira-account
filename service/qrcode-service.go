package service

import (
	"math/rand"
	"time"

	"github.com/ivohutasoit/alira/common"
)

type QrcodeService struct {
}

func (qrcode *QrcodeService) Generate() string {
	b := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = common.Letters[rand.Intn(len(common.Letters))]
	}
	return string(b)
}
