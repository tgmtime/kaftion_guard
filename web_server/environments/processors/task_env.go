package processors

// internal-env-keys
const (
	TaskEnvTag = `task-env-tag`
	//yetki tür sayısına göre uint8, uint32, uint, vb. atama type belirlenir. Karmaşık dinamik izin türleri için argon2id prosedürü işlenir.
	Read uint8 = 1 << iota
	Write
	Run
	Bridge
	Swap
	RWPermType   = Read | Write
	RWSPermType  = Read | Write | Swap
	RWBPermType  = Read | Write | Bridge
	RWSBPermType = Read | Write | Swap | Bridge
)

// internal-env-keys

// external-env-keys

const (
	/*
		belirtilen func üzerinde işlenecek file göre task tanımı yapılır ve bu task tanımı owner
		task tanımında varsa süreç devam edilir.


		Ex: main cbor içersinden çekilecek tanım
		FuncPerm: map[string][string]uint8 => map["func_key"]["file_key"]perm
	*/
	FuncIncludeEnvMapPerm = `func-include-env-map-perm`
	FuncGetEnvPerm        = `func-get-env-perm`
	// IncPathEnvPerm      = `inc-path-env-perm`       //RWPermType
	// IncTaskEnvPerm      = `inc-task-env-perm`       //RWSPermType
	// IncRestEnvPerm      = `inc-rest-env-perm`       //RWBPermType
	// IncFuncEnvPerm      = `inc-func-env-perm`       //RWSBPermType
	// IncFuncErrorEnvPerm = `inc-func-error-env-perm` //RWSPermType
	//func bazında aranan yetkiler
)

// task env external olarak içeri aktarılacak env map içerisinde barınan keys doğrulamak için reference alınacak slice
var TaskEnvKeyRefSlice []string = []string{
	FuncIncludeEnvMapPerm,
	FuncGetEnvPerm,
}

// external-env-keys
