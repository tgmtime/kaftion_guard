package processors

import (
	"path/filepath"
)

// internal-env-keys
const (
	MainPathEnvsPathKey = iota
	SpecificPathKey

	MainPathEnvsPath     = `environments/data`
	SystemPubKeyField    = `system-pub-key.cbor`
	MainEnvMapField      = `main-env.cbor`
	WhitelistEnvMapField = `whitelist-env.cbor`

	ExternalEnvPathKey //external tanımlanan env için

	//sabit anahtarılar static olarak belirtilir.

	PathEnvTag = `path-env-tag`

	DeveloperPubKeyField = `developer-pub-key.cbor`

	OwnerEnvAuthnTokenField = `owner-env-authn-token.cbor`
	//external eklenecek env tanımlandığı alan
	PathEnvMapField      = `path-env.cbor`
	TaskEnvMapField      = `task-env.cbor`
	RestEnvMapField      = `rest-env.cbor`
	FuncEnvMapField      = `func-env.cbor`
	FuncErrorEnvMapField = `func-error-env.cbor`
	//external eklenecek env tanımlandığı alan
	Cbor = `.cbor`
	Json = `.json`
	SH   = `.sh`
)

// internal-env-keys

// external-env-keys
const ()

// path env external olarak içeri aktarılacak env map içerisinde barınan keys doğrulamak için reference alınacak slice
var PathEnvKeyRefSlice []string = []string{}

// external-env-keys
func GetPath(pathKey int, PathFields ...string) (string, error) {
	switch pathKey {
	case MainPathEnvsPathKey: //TODO: gecici internal olarak verildi. tek seferlik imzalı file okumaya göre yapılandırılacak
		return filepath.Join(MainPathEnvsPath, PathFields[0]), nil
	case SpecificPathKey:
		return PathFields[0], nil
	default:
		return "", GetFuncError(InvalidPathKey, nil)
	}
}
