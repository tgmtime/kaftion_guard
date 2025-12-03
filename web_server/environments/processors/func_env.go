package processors

import (
	"strings"
)

// internal-env-keys
const (
	// *******internal func env keys*******
	FuncEnvTag   = `func-env-tag`
	SystemPubKey = `system-pub-key`
	OK           = iota
	SpecificOK
	NotOK
	SpecificNotOK
	SuccessIncluded

	SignTypeED25519
	CheckPubKeyIndex = 0x00
	DefaultSHALength = -1 //default sha uzunluğunda üretim sağlar 32byte

	Exit                      = `ex`
	EndEnvMapField            = `end-env-map-field`
	RandStringKeySpecialChars = `!@#$%^&*()-_=+[]{}|;:'\",.<>?/`
	RandStringKeyChars        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	RandInputMaxCount         = 33
	RandInputMinCount         = 7
	RandSpecialCharMaxCount   = 13

	//*******internal func env keys*******

	//*******external func env keys*******
	HashTypeSHA3_256 uint8 = 0
	HashTypeSHA3_512 uint8 = 1
	HashTypeSHA2_256 uint8 = 2
	//*******external func env keys*******
)

// internal-env-keys

// external-env-keys
const ()

// func env external olarak içeri aktarılacak env map içerisinde barınan keys doğrulamak için reference alınacak slice
var FuncEnvKeyRefSlice []string = []string{}

// external-env-keys

func GetFuncStatus(code int, fields ...string) string {
	switch code {
	case OK:
		return `🟢 success`
	case SpecificOK:
		return `🟢 success ` + strings.Join(fields, " ")
	case NotOK:
		return `⚫ unsuccessful`
	case SpecificNotOK:
		return `⚫ unsuccessful ` + strings.Join(fields, " ")
	case SuccessIncluded:
		return `🟢 successfully included: ` + fields[0]
	default:
		return GetFuncError(InvalidFuncStatusCode, nil).Error()
	}
}
