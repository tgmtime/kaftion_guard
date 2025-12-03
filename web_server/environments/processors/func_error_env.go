package processors

import (
	"errors"
	"fmt"
)

// internal-env-keys
const (
	FuncErrorEnvTag = `func-error-env-tag`

	UnexpectedError = iota
	SystemPathEnvNotFound
	GetAllFieldsRequired
	FileNotFound
	PermissionDenied
	InvalidPathKey
	InvalidFuncStatusCode
	InvalidEnvStatus
	InvalidNewDataStatus
	InactiveDataStatus
	InvalidNewDataExpiresAt
	DataExpiresAt
	InvalidNewDataActiveAt
	DataActiveAt
	InvalidNewDataStatusDescription
	RequiredDataStatusDescription
	InvalidSignatureComponents
	InvalidTaskAuthn
	InvalidAccessData
	InvalidSliceElement
	InvalidSignType
	InvalidEnvMapKeySlice
	InvalidValue
	InvalidEnvMap
	InvalidEnvMapKey
	InvalidMapType
	InvalidMapValueType
	InvalidMapValue
	MissingStatusInfo
	NilValueNotAllowed
	InvalidDataType
	InvalidOwnerKey
	InvalidCID
	InvalidCIDType
	CIDMismatch
	OwnerKeyAlreadyExists
	EnvMapKeyAlreadyExists
	EnvMapKeyNotFound
	EnvMapTypeMismatch
	EnvKeyNotFound
	UnSupportedDataType
	AllFieldsRequired
	AllFieldsRequiredWithInvalidKey
)

// internal-env-keys

// external-env-keys
const ()

// func error env external olarak içeri aktarılacak env map içerisinde barınan keys doğrulamak için reference alınacak slice
var FuncErrorEnvKeyRefSlice []string = []string{}

// external-env-keys

func GetFuncError(code int, err error, fields ...any) error {

	switch code {
	//***dinamic errors***

	//***dinamic errors***

	//*******static errors*******
	case SystemPathEnvNotFound:
		return fmt.Errorf("🔴 system path env file not found")
	case InvalidPathKey:
		return fmt.Errorf("🟡 invalid path key")
	case FileNotFound:
		return fmt.Errorf("🟡 file does not exist at path: %s", fields[0])
	case PermissionDenied:
		return fmt.Errorf("🟡 permission denied for path: %s", fields[0])
	case InvalidNewDataStatus:
		return errors.New("🟡 the status activity information of newly added data cannot be passive")
	case InactiveDataStatus:
		return errors.New("🟡 data status is inactive")
	case InvalidNewDataExpiresAt:
		return errors.New("🟡 expiresAt of newly added data cannot be smaller than now and must be assigned a value of 0 to make it unlimited")
	case DataExpiresAt:
		return errors.New("🟡 data has expired")
	case InvalidNewDataActiveAt:
		return errors.New("🟡 the ActiveAt field for newly created data cannot be in the past or empty")
	case DataActiveAt:
		return errors.New("🟡 data is not yet active")
	case InvalidNewDataStatusDescription:
		return errors.New("🟡 the status Description field for newly created data cannot be empty or invalid")
	case RequiredDataStatusDescription:
		return errors.New("🟡 data status description is required")
	case GetAllFieldsRequired:
		return errors.New("🟡 all fields are required in the system environment data")
	case InvalidFuncStatusCode:
		return errors.New(`🔴 invalid func status code`)
	case InvalidSignatureComponents:
		return errors.New(`🔴 invalid signature components`)
	case InvalidAccessData:
		return errors.New(`🔴 invalid access data`)
	case InvalidTaskAuthn:
		return fmt.Errorf("🔴 invalid task authn: %s", fields[0])
	case InvalidSliceElement:
		return fmt.Errorf("invalid slice element index: %d, error: %v", fields[0], err)
	case InvalidEnvMapKeySlice:
		return fmt.Errorf(`reference env map key slice and input env map key slice slice do not match, 
		env map tag: %s`, fields[0])
	case NilValueNotAllowed:
		return fmt.Errorf("the slice variable cannot be nil: %v", fields[0])
	case UnSupportedDataType:
		return fmt.Errorf("🟡 unsupported data type: %T", fields[0])
	case AllFieldsRequired:
		return errors.New("🟡 all fields are required")
	case AllFieldsRequiredWithInvalidKey:
		return fmt.Errorf("🟡 all fields are required, invalid key/index: %v", fields[0])
	case OwnerKeyAlreadyExists:
		return fmt.Errorf("🟡 public key owner key already exists: %s", fields[0])
	case InvalidOwnerKey:
		return fmt.Errorf("🟡 invalid owner key: %s", fields[0])
	case InvalidDataType:
		return fmt.Errorf("🟡 invalid data type: (type=%T)", fields[0])
	case EnvMapKeyAlreadyExists:
		return fmt.Errorf("🟡 environment map key already exists: %s", fields[0])
	case EnvMapKeyNotFound:
		return fmt.Errorf("🟡 environment map key not found: %s", fields[0])
	case EnvMapTypeMismatch:
		return errors.New("env map type mismatch detected")
	case EnvKeyNotFound:
		return fmt.Errorf("🟡 env key not found in environment map: key=%v (type=%T)", fields[0], fields[0])
	case InvalidValue:
		return fmt.Errorf("🟡 key containing unsupported value: %v", fields[0])
	case InvalidEnvMap:
		return fmt.Errorf("🟡 invalid env map: %v", fields[0])
	case InvalidEnvMapKey:
		return fmt.Errorf("🟡 invalid env map key: %v", fields[0])
	case InvalidMapType:
		return fmt.Errorf("🟡 invalid map type: %v", fields[0])
	case InvalidMapValueType:
		return fmt.Errorf("🟡 invalid or unsupported map value type: %v", fields[0])
	case InvalidMapValue:
		return fmt.Errorf("🟡 invalid map value: %v, error: %v", fields[0], err)
	case InvalidSignType:
		return errors.New("invalid singnature type")
	case MissingStatusInfo:
		return fmt.Errorf("🟡 missing status info: %v", fields[0])
	case InvalidCID:
		return errors.New("Invalid CID: Must be CID v1 with a SHA2-256 multihash. Please ensure compliance with CID standards.")
	case InvalidCIDType:
		return errors.New("invalid CID type")
	case CIDMismatch:
		return errors.New("CID Incompatibility: The provided CIDs do not match.")
	default:
		return errors.New(`🔴 invalid func error code`)
		//*******static errors*******
	}
}
