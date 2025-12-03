package utils

import (
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"time"
	env "web_server/environments/processors"

	cid "github.com/ipfs/go-cid"

	"golang.org/x/crypto/sha3"

	"github.com/fxamacker/cbor/v2"
	"github.com/multiformats/go-multihash"
)

func GenerateNativeBytetoSHA2(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func GenerateNativeBytetoSHA3(data []byte) []byte {
	hash := sha3.New256()
	hash.Write(data)
	return hash.Sum(nil)
}

func GenerateBytetoSHA(data []byte, hashType uint8) ([]byte, error) {
	switch hashType {
	case env.HashTypeSHA3_256:
		return multihash.Sum(data, multihash.SHA3_256, env.DefaultSHALength)
	case env.HashTypeSHA3_512:
		return multihash.Sum(data, multihash.SHA3_512, env.DefaultSHALength)
	case env.HashTypeSHA2_256:
		return multihash.Sum(data, multihash.SHA2_256, env.DefaultSHALength)
	default:
		// Varsayılan olarak SHA3-256 kullanıyoruz
		return multihash.Sum(data, multihash.SHA3_256, env.DefaultSHALength)
	}
}

// string ifadeyi sha3_256 türünden işler ve string çıktısını verir.
func GenerateStringtoSHA(data string, hashType uint8) (string, error) {
	hash, err := GenerateBytetoSHA([]byte(data), hashType)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}

// any türünden herhangi bir datayı cbor ile []byte türüne cevrilerek belirtilen sha türünden işler ve []byte sha çıktısını verir.
func GenerateAnytoSHA(data any, hashType uint8) ([]byte, error) {
	encodeData, err := cbor.Marshal(data)
	if err != nil {
		return nil, env.GetFuncError(env.UnexpectedError, err)
	}

	hash, err := GenerateBytetoSHA(encodeData, hashType)
	if err != nil {
		return nil, env.GetFuncError(env.UnexpectedError, err)
	}
	return hash, nil
}

func ComparisonHashString(inputHash, referenceHash string) bool {

	inputHashBytes, err := hex.DecodeString(inputHash)
	if err != nil {
		return false
	}

	referenceHashBytes, err := hex.DecodeString(referenceHash)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(inputHashBytes, referenceHashBytes) == 1
}

func CompareBase64(base64Str1, base64Str2 string) bool {
	// Base64 decode
	data1, err := base64.StdEncoding.DecodeString(base64Str1)
	if err != nil {
		return false
	}

	data2, err := base64.StdEncoding.DecodeString(base64Str2)
	if err != nil {
		return false
	}

	// Güvenli kıyaslama
	return subtle.ConstantTimeCompare(data1, data2) == 1
}

func ComparisonString(inputStr1, inputStr2 string) bool {
	// String'leri byte dizilerine dönüştür
	input1 := []byte(inputStr1)
	input2 := []byte(inputStr2)

	// Güvenli karşılaştırma yap
	return subtle.ConstantTimeCompare(input1, input2) == 1
}

func ComparisonHash(inputHash, referenceHash []byte) bool {
	return subtle.ConstantTimeCompare(inputHash, referenceHash) == 1
}

func StringSliceToBytes(slice []string) []byte {
	var builder strings.Builder
	for _, s := range slice {
		builder.WriteString(s)
	}
	return []byte(builder.String())
}

// CID'nin geçerliliğini kontrol et
func IsValidCID[T []byte | cid.Cid | string](cidBytes T) error {
	var parsedCID cid.Cid
	var err error

	// Türü kontrol et ve uygun şekilde parse et
	switch v := any(cidBytes).(type) {
	case []byte:
		parsedCID, err = cid.Cast(v)
	case cid.Cid:
		parsedCID = v
	case string:
		parsedCID, err = cid.Decode(v)
	default:
		return env.GetFuncError(env.InvalidCIDType, nil)
	}

	if err != nil {
		return env.GetFuncError(env.InvalidCID, err)
	}

	// Versiyon kontrolü
	if parsedCID.Version() != 1 {
		return env.GetFuncError(env.InvalidCID, err)
	}

	// Codec kontrolü (DagCBOR = 0x71)
	if parsedCID.Type() != cid.DagCBOR {
		return env.GetFuncError(env.InvalidCID, err)
	}

	// Multihash bilgisini al
	mh := parsedCID.Hash()

	// Multihash'ı decode et
	decodedMH, err := multihash.Decode(mh)
	if err != nil {
		return env.GetFuncError(env.InvalidCID, err)
	}

	// SHA2-256 olup olmadığını kontrol et
	if decodedMH.Code != multihash.SHA2_256 {
		return env.GetFuncError(env.InvalidCID, err)
	}

	return nil
}

func StringCIDv1Compare(referenceCID []byte, externalCID string) error {
	// CID'leri parse edip binary formata çevir
	extCID, err := cid.Decode(externalCID)
	if err != nil {
		return env.GetFuncError(env.InvalidCID, err)
	}

	if err := IsValidCID(extCID); err != nil {
		return err
	}

	// Binary formata çevir
	extByteCID := extCID.Bytes()

	// Güvenli karşılaştırma yap (Uzunluk kontrolüne gerek yok)
	if subtle.ConstantTimeCompare(referenceCID, extByteCID) != 1 {
		return env.GetFuncError(env.CIDMismatch, nil)
	}
	return nil
}

func GenerateHashFromKey(data []byte, hashType uint8) ([]byte, error) {
	switch hashType {
	case env.HashTypeSHA3_256:
		return multihash.Sum(data, multihash.SHA3_256, env.DefaultSHALength)
	case env.HashTypeSHA3_512:
		return multihash.Sum(data, multihash.SHA3_512, env.DefaultSHALength)
	case env.HashTypeSHA2_256:
		return multihash.Sum(data, multihash.SHA2_256, env.DefaultSHALength)
	default:
		// Varsayılan olarak SHA3-256 kullanıyoruz
		return multihash.Sum(data, multihash.SHA3_256, env.DefaultSHALength)
	}
}

// default olarak cid byte çıktısını verir
func DatatoCIDv1Byte(data []byte) ([]byte, error) {
	//default olarak multihash.SHA2_256
	hash, err := GenerateHashFromKey(data, env.HashTypeSHA2_256)
	if err != nil {
		return nil, err
	}
	return cid.NewCidV1(cid.DagCBOR, hash).Bytes(), nil
}

func RandomCharFromSet(charSet string) (byte, error) {
	index, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(charSet))))
	if err != nil {
		return 0, err
	}
	return charSet[index.Int64()], nil
}

func GenerateRandomStringKey(minLength, maxLength, specialCharCount int) (string, error) {
	// Anahtar uzunluğunu belirle
	lengthRange := maxLength - minLength + 1
	keyLength, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(lengthRange)))
	if err != nil {
		return "", err
	}
	keyLength = keyLength.Add(keyLength, big.NewInt(int64(minLength)))

	// Belirtilen sayıda özel karakter ekle
	var keyBuilder strings.Builder
	for i := 0; i < specialCharCount; i++ {
		char, err := RandomCharFromSet(env.RandStringKeySpecialChars)
		if err != nil {
			return "", err
		}
		keyBuilder.WriteByte(char)
	}

	// Kalan karakterleri büyük/küçük harf ve sayılardan rastgele seç
	remainingLength := int(keyLength.Int64()) - specialCharCount
	for i := 0; i < remainingLength; i++ {
		char, err := RandomCharFromSet(env.RandStringKeyChars)
		if err != nil {
			return "", err
		}
		keyBuilder.WriteByte(char)
	}

	// Karakterleri karıştır
	key := []rune(keyBuilder.String())

	// Yerel ve güvenli bir RNG oluştur
	seed, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", err
	}
	rng := rand.New(rand.NewSource(seed.Int64())) // Yerel RNG
	rng.Shuffle(len(key), func(i, j int) {
		key[i], key[j] = key[j], key[i]
	})

	return string(key), nil
}

func GenerateRandStringKeyRandInput() (minLength, maxLength, specialCharCount int) {
	var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Min ve max uzunluk için rastgele değerler belirle (13 ile 33 arasında)
	minLength = rng.Intn(21) + 13        // En az 13 karakter
	maxLength = minLength + rng.Intn(21) // Maksimum minLength + 0-21 arası rastgele

	// Özel karakter sayısını belirle (maxLength'ten büyük olamaz ve minLength'in 3'te biri olmalı)
	specialCharCount = rng.Intn(minLength/3 + 1) // minLength'in 3'te birini geçmemesi için

	return minLength, maxLength, specialCharCount
}

func GenerateAccessToken()
