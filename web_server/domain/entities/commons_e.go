package entities

type StatusData struct {
	Status      bool   `cbor:"1,keyasint" yaml:"status"`
	CreatedAt   int64  `cbor:"2,keyasint" yaml:"created-at"`
	ActiveAt    int64  `cbor:"3,keyasint" yaml:"active-at"`
	ExpiresAt   int64  `cbor:"4,keyasint" yaml:"expires-at"`
	UpdatedAt   int64  `cbor:"5,keyasint" yaml:"updated-at"`
	Description string `cbor:"6,keyasint" yaml:"description"`
}

type FieldData struct {
	CID         []byte     `cbor:"1,keyasint"`
	OtherURL    string     `cbor:"2,keyasint"`
	Description string     `cbor:"3,keyasint"`
	StatusInfos StatusData `cbor:"4,keyasint"`
}

type SpecificData struct {
	Permission string    `cbor:"1,keyasint"`
	FieldInfos FieldData `cbor:"2,keyasint"`
}

