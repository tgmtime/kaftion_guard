package processors

import (
	"reflect"
	"sync"
	"unsafe"
	e "web_server/domain/entities"
)

// internal-env-keys
const (
	MainEnvTag = `main-env-tag`
)

// internal-env-keys

// external-env-keys
const (
	System    = `system`
	
	//kullanılanlar

	DevKeyAccessInfosKey = `dev-access-infos` // bu değişken altında map["owner_name"] => (access_data_path, StatusInfo)
	AccessDataPath       = `access-data-path`
	ExternalEnvBasePath  = `external-env-base-path`
	FuncPerm             = `func-perm`
	OwnerWhitelist       = `owner-whitelist`
	//system funcs izinleri

)

// path env external olarak içeri aktarılacak env map içerisinde barınan keys doğrulamak için reference alınacak slice
var MainEnvKeyRefSlice []string = []string{
	FuncPerm,
	OwnerWhitelist,

	AccessDataPath,

	ExternalEnvBasePath,
}

// external-env-keys

// Genel ortam değişkenlerini saklamak için bir yapı
// var envMaps = map[string]*unsafe.Pointer{
// 	MainEnvMapField:      &mainEnvs,
// 	PathEnvMapField:      &pathEnvs, //new(unsafe.Pointer),
// 	TaskEnvMapField:      &taskEnvs,
// 	RestEnvMapField:      &restEnvs,
// 	FuncEnvMapField:      &funcEnvs,
// 	FuncErrorEnvMapField: &funcErrorEnvs,
// 	WhitelistEnvMapField: &whitelistEnvs,
// }

// ****general env map operations****
// Ortam değişkenleri ve kilitler
var (
	envMaps     sync.Map // map[string]EnvMapData[K,V] - Ana thread-safe harita
	envMapLocks sync.Map // map[string]*sync.RWMutex - Her envMapKey için özel kilit
)

// getEnvLock belirli bir envMapKey için kilidi atomik olarak alır veya oluşturur
func getEnvLock(envMapKey string) *sync.RWMutex {
	lock, _ := envMapLocks.LoadOrStore(envMapKey, &sync.RWMutex{})
	return lock.(*sync.RWMutex)
}

// cloneMap generic map kopyalama fonksiyonu (Thread-safe değil, caller sorumlu)
// cloneMap deep copy
func cloneMap[K comparable, V any](src map[K]V) map[K]V {
	if src == nil {
		return nil
	}

	dst := make(map[K]V, len(src))
	for k, v := range src {
		dst[deepCopy(k).(K)] = deepCopy(v).(V)
	}
	return dst
}

// isPrimitive checks if the kind is a basic type that can be shallow copied
func isPrimitive(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String, reflect.Bool:
		return true
	default:
		return false
	}
}

// deepCopy performs an optimized deep copy using type assertions and reflection
func deepCopy(src any) any {
	if src == nil {
		return nil
	}

	// Handle primitive types and safe copies first
	switch v := src.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, complex64, complex128,
		string, bool:
		return v

	case []byte:
		c := make([]byte, len(v))
		copy(c, v)
		return c

	case *int:
		if v == nil {
			return (*int)(nil)
		}
		cpy := *v
		return &cpy

		// Add other pointer and slice types as needed...
	}

	// Handle complex types with reflection
	val := reflect.ValueOf(src)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return reflect.New(val.Type()).Interface()
		}
		cpy := reflect.New(val.Type().Elem())
		cpy.Elem().Set(reflect.ValueOf(deepCopy(val.Elem().Interface())))
		return cpy.Interface()

	case reflect.Slice:
		if val.IsNil() {
			return reflect.Zero(val.Type()).Interface()
		}
		// Optimize primitive slices with reflect.Copy
		if elemKind := val.Type().Elem().Kind(); isPrimitive(elemKind) {
			cpy := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
			reflect.Copy(cpy, val)
			return cpy.Interface()
		}
		// Deep copy for complex slices
		cpy := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			cpy.Index(i).Set(reflect.ValueOf(deepCopy(val.Index(i).Interface())))
		}
		return cpy.Interface()

	case reflect.Map:
		if val.IsNil() {
			return reflect.Zero(val.Type()).Interface()
		}
		cpy := reflect.MakeMapWithSize(val.Type(), val.Len())
		for _, key := range val.MapKeys() {
			origValue := val.MapIndex(key)
			cpy.SetMapIndex(
				reflect.ValueOf(deepCopy(key.Interface())),
				reflect.ValueOf(deepCopy(origValue.Interface())),
			)
		}
		return cpy.Interface()

	case reflect.Struct:
		cpy := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			if !cpy.Field(i).CanSet() {
				continue
			}
			fieldCopy := deepCopy(val.Field(i).Interface())
			cpy.Field(i).Set(reflect.ValueOf(fieldCopy))
		}
		return cpy.Interface()

	case reflect.Interface:
		if val.IsNil() {
			return nil
		}
		return deepCopy(val.Elem().Interface())

	default:
		if val.CanInterface() {
			return copyUnexported(val)
		}
		return src
	}
}

// copyUnexported carefully copies unexported fields
func copyUnexported(original reflect.Value) any {
	if !original.IsValid() {
		return nil
	}

	cpy := reflect.New(original.Type()).Elem()
	size := original.Type().Size()

	// Use unsafe to copy underlying memory
	origPtr := unsafe.Pointer(original.UnsafeAddr())
	cpyPtr := unsafe.Pointer(cpy.UnsafeAddr())
	copy((*[1 << 30]byte)(cpyPtr)[:size], (*[1 << 30]byte)(origPtr)[:size])

	return cpy.Interface()
}

// SetNewEnvMap yeni bir ortam haritası oluşturur (Tamamen thread-safe)
func SetNewEnvMap[K comparable, V any](envMapKey string, input e.EnvMapData[K, V]) error {
	lock := getEnvLock(envMapKey)
	lock.Lock()
	defer lock.Unlock()

	if _, exists := envMaps.Load(envMapKey); exists {
		return GetFuncError(EnvMapKeyAlreadyExists, nil, envMapKey)
	}

	clonedEnvInfos := cloneMap(input.EnvInfos)
	newData := e.EnvMapData[K, V]{
		EnvInfos:    clonedEnvInfos,
		StatusInfos: input.StatusInfos,
	}
	envMaps.Store(envMapKey, newData)
	return nil
}

// UpdateEnvMap varolan bir ortam haritasını atomik olarak günceller
func UpdateEnvMap[K comparable, V any](envMapKey string, input e.EnvMapData[K, V]) error {
	lock := getEnvLock(envMapKey)
	lock.Lock()
	defer lock.Unlock()

	current, exists := envMaps.Load(envMapKey)
	if !exists {
		return GetFuncError(EnvMapKeyNotFound, nil, envMapKey)
	}

	if _, valid := current.(e.EnvMapData[K, V]); !valid {
		return GetFuncError(EnvMapTypeMismatch, nil)
	}

	clonedEnvInfos := cloneMap(input.EnvInfos)
	newData := e.EnvMapData[K, V]{
		EnvInfos:    clonedEnvInfos,
		StatusInfos: input.StatusInfos,
	}
	envMaps.Store(envMapKey, newData)
	return nil
}

func GetEnvMap[K comparable, V any](envMapKey string) (e.EnvMapData[K, V], error) {
	lock := getEnvLock(envMapKey)
	lock.RLock()
	defer lock.RUnlock()

	rawData, exists := envMaps.Load(envMapKey)
	if !exists {
		return e.EnvMapData[K, V]{}, GetFuncError(EnvMapKeyNotFound, nil, envMapKey)
	}

	typedData, valid := rawData.(e.EnvMapData[K, V])
	if !valid {
		return e.EnvMapData[K, V]{}, GetFuncError(EnvMapTypeMismatch, nil)
	}

	// Clone EnvInfos to prevent external mutations
	clonedEnvInfos := cloneMap(typedData.EnvInfos)

	// If StatusInfos is mutable, add deep-cloning logic here
	return e.EnvMapData[K, V]{
		EnvInfos:    clonedEnvInfos,
		StatusInfos: typedData.StatusInfos, // Ensure immutability or clone
	}, nil
}

// GetEnv belirli bir ortam değerini thread-safe şekilde getirir
func GetEnv[K comparable, V any](envMapKey string, key K) (V, error) {
	var zero V
	lock := getEnvLock(envMapKey)
	lock.RLock()
	defer lock.RUnlock()

	rawData, exists := envMaps.Load(envMapKey)
	if !exists {
		return zero, GetFuncError(EnvMapKeyNotFound, nil)
	}

	typedData, valid := rawData.(e.EnvMapData[K, V])
	if !valid {
		return zero, GetFuncError(EnvMapTypeMismatch, nil)
	}

	value, found := typedData.EnvInfos[key]
	if !found {
		return zero, GetFuncError(EnvKeyNotFound, nil, key)
	}
	return value, nil
}

// DeleteEnvMap bir ortam haritasını atomik olarak siler
func DeleteEnvMap(envMapKey string) {
	lock := getEnvLock(envMapKey)
	lock.Lock()
	defer lock.Unlock()

	envMaps.Delete(envMapKey)
	envMapLocks.Delete(envMapKey)
}

// ****general env map operations****

// ****Pubkey operations****
var pubKeyEnvs sync.Map // *PubKeyData saklar

func SetNewPubKey(ownerKey string, data e.PubKeyData) error {
	newData := clonePubKeyData(data)

	// Atomic check-and-set
	if _, loaded := pubKeyEnvs.LoadOrStore(ownerKey, newData); loaded {
		return GetFuncError(OwnerKeyAlreadyExists, nil, ownerKey)
	}
	return nil
}

func UpdatePubKey(ownerKey string, data e.PubKeyData) error {
	newData := clonePubKeyData(data)

	// Optimistic locking with retry
	for {
		oldVal, exists := pubKeyEnvs.Load(ownerKey)
		if !exists {
			return GetFuncError(InvalidOwnerKey, nil, ownerKey)
		}

		// Compare-and-swap pattern
		if pubKeyEnvs.CompareAndSwap(ownerKey, oldVal, newData) {
			return nil
		}
		// Retry if value changed concurrently
	}
}

func GetPubKey(ownerKey string) (*e.PubKeyData, error) {
	value, exists := pubKeyEnvs.Load(ownerKey)
	if !exists {
		return nil, GetFuncError(InvalidOwnerKey, nil, ownerKey)
	}

	// Type safety check
	dataPtr, ok := value.(*e.PubKeyData)
	if !ok {
		pubKeyEnvs.Delete(ownerKey) // Geçersiz veriyi temizle
		return nil, GetFuncError(InvalidDataType, nil, value)
	}

	// İç veriyi korumak için klonlanmış kopyayı döndür
	return clonePubKeyData(*dataPtr), nil
}

func DeletePubKeyEnv(ownerKey string) {
	pubKeyEnvs.Delete(ownerKey)
}

// Yardımcı fonksiyonlar
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func clonePubKeyData(data e.PubKeyData) *e.PubKeyData {
	return &e.PubKeyData{
		PubKey:     cloneBytes(data.PubKey),
		StatusInfo: data.StatusInfo, // StatusData'nın değer tipi olduğu varsayılıyor
	}
}

// ****Pubkey operations****
