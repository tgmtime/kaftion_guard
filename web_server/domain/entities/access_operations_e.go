package entities

type VerifySignInput[T any] struct {
	SignType  int
	PublicKey []byte
	Signed    []byte
	Data      T
}
