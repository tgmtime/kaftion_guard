package entities

type GetInput[T any] struct {
	PathKey    int      `cbor:"1,keyasint"`
	PathFields []string `cbor:"2,keyasint"`
	Data       *T       `cbor:"3,keyasint"`
}

type EnvMapChainData struct {
	NextEnvMap     string
	EnvKeyRefSlice []string
}

type IncludeExternalEnvInput struct {
	OwnerWhitelistKeyInfo OwnerWhitelistData
	StartEnvMapField      string
	EnvMapChainInfos      map[string]EnvMapChainData
	// IncludeEnvMapInfos map[string]IncludeEnvData

}

type WebServerConfigEnv struct {
	Port         string
	AllowMethods []string
	AllowHeaders []string
	AllowOrigins []string
}

// *******genel env file formatı*******
type EnvData[V any] struct {
	Value        V            `cbor:"1,keyasint"`
	SpecificInfo SpecificData `cbor:"2,keyasint"`
	StatusInfo   StatusData   `cbor:"3,keyasint"`
}

type EnvMapData[K comparable, V any] struct {
	EnvInfos     map[K]V      `cbor:"1,keyasint"`
	SpecificInfo SpecificData `cbor:"2,keyasint"`
	StatusInfos  StatusData   `cbor:"3,keyasint"`
}

type EnvFileData[K comparable, V any] struct {
	EnvMapInfos    EnvMapData[K, V] `cbor:"1,keyasint"`
	SignatureInfos SignatureData    `cbor:"2,keyasint"`
}

// *******genel env file formatı*******

type ValidateEnvMapInput[V any] struct {
	ReferenceEnvTag   string
	ReferenceEnvSlice []string
	ExternalEnvMap    map[string]EnvData[V]
}

// *******pubkey*******
type PubKeyData struct {
	PubKey     []byte     `cbor:"1,keyasint"`
	StatusInfo StatusData `cbor:"2,keyasint"`
}

// *******pubkey*******

//*******Whitelist*******

type PathData struct {
	URI        string     `cbor:"1,keyasint"`
	StatusInfo StatusData `cbor:"2,keyasint"`
}

// ExternalPathInfos: file_name_key: sha2(file_name) => map[file_name_key] => uri&status 78.189.17.129 @Eyn7?ek{=/(XfT&DZ3[)s*v^
type WhitelistOwnerData struct {
	PubKeyDataURI string     `cbor:"1,keyasint"` //owner pubkey barındıran file uri
	AccessDataCID []byte     `cbor:"2,keyasint"` //owner yetkilerini barındıran access data cid verir. sisteme ilk yüklenirken []byte türünden yükleme gercekleştirilir.
	StatusInfos   StatusData `cbor:"3,keyasint"` //WhitelistOwnerData struct status bilgisini verir.
}

// type WhitelistData struct {
// 	WhitelistOwnerInfos map[string]WhitelistOwnerData `cbor:"1,keyasint"`
// 	StatusInfos         StatusData                    `cbor:"2,keyasint"`
// }

/*
- owner authn config file data
- WhitelistInfos: sistem üzerinde işlen yapan owner(developer) ve bu developer ait olan
status bilgilerini tutar.
- WhitelistInfos: içerisindeki WhitelistOwnerInfos owner_whitelist_access_key_info: sha2(password+account) =>  map[owner_whitelist_access_key_info]  => WhitelistOwnerData
- sistem içerisine belirtilen accounts bilgisinin bilgileri içerisindeki account adına ulaşmak
istedikleri external sub uri bilgileri ve doğrulama için 7sn geçerli olan timestamp bilgisi ve bu bilgilerin imzası verilir.
verilen bu bilgileri sistem içerisindeki belirtilen account pubkey ile imza doğrulaması gerecekleştirilir ve başarılı olursa
sistem belirtilen env yüklemesi gercekleşir.

env yükleme mekanizması

	{
		WhitelistInfos EnvMapData[string, WhitelistOwnerData] `cbor:"1,keyasint"`
		owner_whitelist_access_key_info: map_key: sha2(password+account)
		timastamp: int64
	}
*/
type SystemWhiteListData[K comparable, V any] struct {
	WhitelistInfos EnvMapData[K, V] `cbor:"1,keyasint"`
	SignatureInfos SignatureData    `cbor:"2,keyasint"` //sistem tarafından oluşturulur.
}

//*******Whitelist*******

// sistem setup config
type SetupConfigData struct {
	SystemTag         string          `cbor:"1,keyasint" yaml:"system-tag"`
	SystemVersion     string          `cbor:"2,keyasint" yaml:"system-version"`
	SystemPubKeyPath  string          `cbor:"3,keyasint" yaml:"system-pub-key-path"`
	SystemName        string          `cbor:"4,keyasint" yaml:"system-name"`
	SystemDescription string          `cbor:"5,keyasint" yaml:"system-description"`
	ConfigPath        string          `cbor:"6,keyasint" yaml:"config-path"`
	LogLevel          string          `cbor:"7,keyasint" yaml:"log-level"`
	EnableDebug       bool            `cbor:"8,keyasint" yaml:"enable-debug"`
	MaxConnections    int             `cbor:"9,keyasint" yaml:"max-connections"`
	TimeoutSeconds    int             `cbor:"10,keyasint" yaml:"timeout-seconds"`
	DatabaseURL       string          `cbor:"11,keyasint" yaml:"database-url"`
	APIEndpoint       string          `cbor:"12,keyasint" yaml:"api-endpoint"`
	AllowedIPs        []string        `cbor:"13,keyasint" yaml:"allowed-ips"`
	FeatureFlags      map[string]bool `cbor:"14,keyasint" yaml:"feature-flags"`
}

type SystemConfigData struct {
	SetupConfigInfo SetupConfigData `yaml:"system-config-info"`
	StatusInfo      StatusData      `yaml:"status-info"`
}
