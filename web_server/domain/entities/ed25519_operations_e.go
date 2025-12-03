package entities

type VerifySignED25519Input struct {
	PublicKey []byte
	Signed    []byte
	Data      []byte
}

// ********owner access data********
type SignatureData struct {
	SignedBy  string `cbor:"1,keyasint"`
	Signature []byte `cbor:"2,keyasint"`
}

type PermissionData struct {
	PermType    uint8      `cbor:"1,keyasint"` //bridge, read, write, swap, vb.
	StatusInfos StatusData `cbor:"2,keyasint"` //operation türünün status bilgisini tutar.
}

/*
PermInfos: permission_key(sha2 yada direk isimlendirme tarzında permission key oluşturulur)
map[permission_key] => PermType/StatusInfo
*/
type AuthnData struct {
	PermInfos   map[string]PermissionData `cbor:"1,keyasint"`
	FieldInfo   FieldData                 `cbor:"2,keyasint"`
	StatusInfos StatusData                `cbor:"3,keyasint"` //access data status bilgisini tutar.
}

type AccessData struct {
	AuthnInfos          AuthnData     `cbor:"1,keyasint"`
	TaskInfosSignInfos  SignatureData `cbor:"2,keyasint"`
	AuthnInfosSignInfos SignatureData `cbor:"3,keyasint"`
}

// ********owner access data********

// ********owner whitelist üzerindeki etkileşim datası********
type AccessKeyData struct {
	WhitelistKey  string     `cbor:"1,keyasint"` //owner whitelist key bilgisini verir.WhitelistKey: sha2(password+account)
	AccessDataCID []byte     `cbor:"2,keyasint"` //owner yetkilerini barındıran access data cid verir
	StatusInfos   StatusData `cbor:"3,keyasint"`
}

type WhitelistAccessData struct {
	AccessKeyInfos AccessKeyData `cbor:"1,keyasint"`
	SignatureInfos SignatureData `cbor:"2,keyasint"` //owner tarafından üretilen imza
}

// ********owner whitelist üzerindeki etkileşim datası********

// ********access token********
type TokenData struct {
	AccessToken  []byte     `cbor:"1,keyasint"` //strcut{owner_whitelist_key+refresh_token+status_data}
	RefreshToken []byte     `cbor:"2,keyasint"` //struct{owner_whitelist_key+status_data}
	FieldInfo    FieldData  `cbor:"3,keyasint"`
	StatusInfo   StatusData `cbor:"4,keyasint"`
}

// ********access token********
