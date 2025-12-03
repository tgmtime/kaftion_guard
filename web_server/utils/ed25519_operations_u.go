package utils

import (
	"crypto/ed25519"
	e "web_server/domain/entities"
	env "web_server/environments/processors"
)

func VerifySignED25519(input e.VerifySignED25519Input) error {
	if len(input.PublicKey) != ed25519.PublicKeySize ||
		!ed25519.Verify(input.PublicKey, input.Data, input.Signed) {
		return env.GetFuncError(env.InvalidSignatureComponents, nil)
	}
	return nil
}
