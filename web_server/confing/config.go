package config

import (
	"fmt"
	"os"
	"path/filepath"
	a "web_server/domain/abstractions"
	e "web_server/domain/entities"
	env "web_server/environments/processors"
	u "web_server/utils"
	v "web_server/validations"

	"github.com/fsnotify/fsnotify"
	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

const (
	SystemEnvOwner = "system"
	environment    = "environment"

	WebServerConfigEnv uint8 = iota
	SystemConfigEnv
	// DBConfigEnv
)

func InitWebServerConfig() {

}

func InitSystemConfig() {

}

type FileEngine[T any] struct {
	//access durum kontrollü işlemlerde geçerlidir.(Owner: O , Access: A , Data: D)
	Owner      string   //işlemi yapıcak owner içerir.
	OADPathKey int      // işlemi yapıcak owner access data bulunduğu path key içerir.
	OADFields  []string // işlemi yapıcak owner access data fields içerir.
}

// derleme zamanı file engine interface check
var _ a.IFileEngine[struct{}] = (*FileEngine[struct{}])(nil)

func getWhitelistOwnerData(ownerWhitelistKey string) (e.WhitelistOwnerData, error) {
	//owner ait olan whitelist üzerindeki data getirilir.
	data, err := env.GetEnv[string, e.WhitelistOwnerData](env.WhitelistEnvMapField, ownerWhitelistKey)
	if err != nil {
		return e.WhitelistOwnerData{}, err
	}

	//owner ait olan whitelist datasının status bilgisi kontrol edilir.
	if err := u.CheckDataStatusInfos(&e.CheckDataStatusInfosInput{
		Status:      data.StatusInfos.Status,
		ActiveAt:    data.StatusInfos.ActiveAt,
		ExpiresAt:   data.StatusInfos.ExpiresAt,
		Description: data.StatusInfos.Description,
	}); err != nil {
		return e.WhitelistOwnerData{}, err
	}

	return data, nil
}

// owner reference belirtilen işlemleri yapma yetkisine sahip mi kontrol edilir.
func (j *FileEngine[T]) IFAccessOperation(input e.WhitelistAccessData) error {
	/*
		1. dışarıdan gelen herhangi bir istek için öncesinde whitelist doğrulaması yapılması gerekir.
			1.1. kişinin verdiği WhitelistAccessData formatındaki .cbor datası içerisinde belirttiği whitelist key ile sorgusu yapılır.
			1.2. kişinin belirtiği whitelist key mevcut ise whitelist belirtiği pub key ile isteği gönderen kişinin request datasının
			imzası ve status bilgisi sorugulanır.

				yukarıdaki 1.1. ve 1.2. işlemler login gibi değerlendirilebilir işlem sonucunda kişi login işlemi yapmak isterse geçiçi
			süreliğine token verilir ve sisteme sağladığı access data bilgisiyle kaydedilir. Her istekte belli status bilgisine göre süreç 
			devam edilir. token süresi biterse refresh token bilgisiyle devam edilir ve yeni token iletilir.

		2. herhangi bir yükleme okuma vb. bir işlem yapacaksa yetki durumu hiyerarşik olarak kontrol edilir.
	*/

	//owner ait whitelist datası varlığı kontrol edilir ve getirilir. Süreç doğrulamasını gerçekleştirmek için
	whitelistOwnerData, err := getWhitelistOwnerData(input.AccessKeyInfos.WhitelistKey)
	if err != nil {
		return err
	}

	//owner whitelist datası içerisinde belirtilen path üzerinden owner ait olan pubkey getirilir. Getirilme nedeni owner request olarak gönderdiği WhitelistAccessData doğruluğunu kontrol etmek
	ownerPubKeyData, err := getPubKey(env.SpecificPathKey, whitelistOwnerData.PubKeyDataURI)
	if err != nil {
		return err
	}

	/*
	- sonraki steplerde owner istekleri için pubkey sisteme kaydedilir 
	PublicKey set ederken kullanılacak key üretimi:  WhitelistAccessData içerisindeki AccessKeyInfos hash cevrilir.
	sha2(AccessKeyInfos) => public
	- sistem herhangi bir restart durumunda bu bilgi gideceği için sistem imzası içeren bir key olması gerekir.
	bu yüzden WhitelistAccessData içerisindeki AccessKeyInfos sistem tarafından imzalanır. 

	*/
	// if err := env.SetNewPubKey() => e.PubKeyData

	//owner request ile göndermiş olduğu datanın imza kontrolü gercekleştirilir.
	if err := u.VerifySign(e.VerifySignInput[e.AccessKeyData]{
		SignType:  env.SignTypeED25519,
		PublicKey: ownerPubKeyData.PubKey,
		Signed:    input.SignatureInfos.Signature,
		Data:      input.AccessKeyInfos,
	}); err != nil {
		return err
	}

	//bu kısımda kaldın
	//whitelist owner data access data cid ile request ile gönderilen whitelist access data bulunan access data cid eşleşme durumu kontrol edilir.
	if err := u.StringCIDv1Compare(whitelistOwnerData.AccessDataCID, input.AccessKeyInfos.AccessDataCID); err != nil {
		return err
	}

	// var accessFileData T
	// if err := j.IFGet(e.GetInput[T]{
	// 	PathKey:    j.OADPathKey,
	// 	PathFields: j.OADFields,
	// 	Data:       &accessFileData,
	// }); err != nil {
	// 	return err
	// }

	// //e.AccessData tür kontrolü yaparak uygun formatta işlenir.
	// accessData, ok := any(accessFileData).(e.AccessData)
	// if !ok {
	// 	return env.GetFuncError(env.InvalidAccessData, nil)
	// }

	// //access data imzasını kontrol etmek için developer pub key çağrılır.
	// devPubKeyInfo, err := getPubKey(env.DeveloperPubKeyField)
	// if err != nil {
	// 	return err
	// }

	// //- access data içerisindeki developer(owner) tarafından oluşturulan AuthnInfos bilgisinin imza kontrolü sağlanır.
	// if err := u.VerifySign(e.VerifySignInput[e.AuthnData]{
	// 	SignType:  env.SignTypeED25519,
	// 	PublicKey: devPubKeyInfo.PubKey,
	// 	Signed:    accessData.AuthnInfosSignInfo.Signature,
	// 	Data:      accessData.AuthnInfos,
	// }); err != nil {
	// 	return err
	// }

	// //access data imzasını kontrol etmek için developer pub key çağrılır.
	// systemPubKeyInfo, err := getPubKey(env.SpecificPathKey, devPubKeyPath.Value.(string))
	// if err != nil {
	// 	return err
	// }

	// //access data içerisindeki system tarafından oluşturulan TaskInfos bilgisinin imza kontrolü sağlanır.
	// if err := u.VerifySign(e.VerifySignInput[map[string]e.TaskData]{
	// 	SignType:  env.SignTypeED25519,
	// 	PublicKey: systemPubKeyInfo.PubKey,
	// 	Signed:    accessData.TaskInfosSignInfo.Signature,
	// 	Data:      accessData.AuthnInfos.TaskInfos,
	// }); err != nil {
	// 	return err
	// }

	// //access data status bilgisi kontrol edilir.
	// if err := u.CheckDataStatusInfos(&e.CheckDataStatusInfosInput{
	// 	Status:      accessData.AuthnInfos.StatusInfos.Status,
	// 	ActiveAt:    accessData.AuthnInfos.StatusInfos.ActiveAt,
	// 	ExpiresAt:   accessData.AuthnInfos.StatusInfos.ExpiresAt,
	// 	Description: accessData.AuthnInfos.StatusInfos.Description,
	// }); err != nil {
	// 	return err
	// }

	// //authn kontrolü.
	// if err := u.CheckAuthn(accessData.AuthnInfos.TaskInfos, referenceTasks); err != nil {
	// 	return err
	// }
	return nil
}

func (j *FileEngine[T]) IFGetRootFilePath(pathKey int, pathFields ...string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	path, err := env.GetPath(pathKey, pathFields...)
	if err != nil {
		return path, err
	}

	filePath := filepath.Join(dir, path)
	return filePath, nil
}

func (j *FileEngine[T]) IFExists(pathKey int, pathFields ...string) error {
	fullPath, err := j.IFGetRootFilePath(pathKey, pathFields...)
	if err != nil {
		return err
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return env.GetFuncError(env.FileNotFound, nil, fullPath)
		case os.IsPermission(err):
			return env.GetFuncError(env.PermissionDenied, nil, fullPath)
		default:
			return env.GetFuncError(env.UnexpectedError, err)
		}
	}

	return nil
}

func (j *FileEngine[T]) IFGet(input e.GetInput[T]) error {
	fullPath, err := j.IFGetRootFilePath(input.PathKey, input.PathFields...)
	if err != nil {
		return err
	}

	decodedData, err := os.ReadFile(fullPath)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return env.GetFuncError(env.FileNotFound, nil, fullPath)
		case os.IsPermission(err):
			return env.GetFuncError(env.PermissionDenied, nil, fullPath)
		default:
			return env.GetFuncError(env.UnexpectedError, err)
		}
	}

	if err := cbor.Unmarshal(decodedData, input.Data); err != nil {
		return env.GetFuncError(env.UnexpectedError, err)
	}
	return nil
}

func loadYamlConfig[T any](filePath string, config *T) error {

	// os.ReadFile ile dosya oku
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	// YAML parse et
	err = yaml.Unmarshal(data, config)
	return nil
}

func getPubKey(pathKey int, pathFields ...string) (*e.PubKeyData, error) {
	var pubKeyEngine *FileEngine[e.PubKeyData] = &FileEngine[e.PubKeyData]{Owner: env.System}
	//gönderilen pub key length olarak garantiye almak ve bellek tüketimi için 33 verilir. Son index kontrol edilir.
	pubKeyData := &e.PubKeyData{}
	if err := pubKeyEngine.IFGet(e.GetInput[e.PubKeyData]{
		PathKey:    pathKey,
		PathFields: pathFields,
		Data:       pubKeyData,
	}); err != nil {
		return nil, err
	}

	//sorgulana public key var ise status durumu kontrol edilir.
	if err := u.CheckDataStatusInfos(&e.CheckDataStatusInfosInput{
		Status:      pubKeyData.StatusInfo.Status,
		ActiveAt:    pubKeyData.StatusInfo.ActiveAt,
		ExpiresAt:   pubKeyData.StatusInfo.ExpiresAt,
		Description: pubKeyData.StatusInfo.Description,
	}); err != nil {
		return nil, err
	}

	return pubKeyData, nil
}

func incSysWhitelist(sysPubKeyData *e.PubKeyData) error {
	// işlem yapacak whitelist kullancılarının yüklendiği kısım
	var sysWhitelistEng *FileEngine[e.SystemWhiteListData[string, e.WhitelistOwnerData]] = &FileEngine[e.SystemWhiteListData[string, e.WhitelistOwnerData]]{Owner: env.System}
	// sistemin base path bilgisi alınır.
	sysWhiteListData := &e.SystemWhiteListData[string, e.WhitelistOwnerData]{}

	if err := sysWhitelistEng.IFGet(e.GetInput[e.SystemWhiteListData[string, e.WhitelistOwnerData]]{
		PathKey:    env.MainPathEnvsPathKey,
		PathFields: []string{env.WhitelistEnvMapField},
		Data:       sysWhiteListData,
	}); err != nil {
		return err
	}

	//alınan system whitelist datasının status bilgisi kontrol edilir
	if err := u.CheckNewDataStatusInfos(&e.CheckDataStatusInfosInput{
		Status:      sysWhiteListData.WhitelistInfos.StatusInfos.Status,
		ActiveAt:    sysWhiteListData.WhitelistInfos.StatusInfos.ActiveAt,
		ExpiresAt:   sysWhiteListData.WhitelistInfos.StatusInfos.ExpiresAt,
		Description: sysWhiteListData.WhitelistInfos.StatusInfos.Description,
	}); err != nil {
		return err
	}

	// system whitelis data imza kontrolü
	if err := u.VerifySign(e.VerifySignInput[e.EnvMapData[string, e.WhitelistOwnerData]]{
		SignType:  env.SignTypeED25519,
		PublicKey: sysPubKeyData.PubKey,
		Signed:    sysWhiteListData.SignatureInfos.Signature,
		Data:      sysWhiteListData.WhitelistInfos,
	}); err != nil {
		return err
	}

	// system whitelist bütün süreçler olumlu olursa WhitelistEnvMapField key ile sisteme yüklenir.
	if err := env.SetNewEnvMap[string, e.WhitelistOwnerData](env.WhitelistEnvMapField, sysWhiteListData.WhitelistInfos); err != nil {
		return err
	}
	return nil
}

func incMainEnv(sysPubKeyData *e.PubKeyData) error {
	// sistemin ana env yüklenmesi
	var mainEnvEng *FileEngine[e.EnvFileData[string, e.EnvData[[]byte]]] = &FileEngine[e.EnvFileData[string, e.EnvData[[]byte]]]{Owner: env.System}
	mainEnvfileData := &e.EnvFileData[string, e.EnvData[[]byte]]{}
	if err := mainEnvEng.IFGet(e.GetInput[e.EnvFileData[string, e.EnvData[[]byte]]]{
		PathKey:    env.MainPathEnvsPathKey,
		PathFields: []string{env.MainEnvMapField},
		Data:       mainEnvfileData,
	}); err != nil {
		return err
	}

	//main env data imza kontrolü
	if err := u.VerifySign(e.VerifySignInput[e.EnvMapData[string, e.EnvData[[]byte]]]{
		SignType:  env.SignTypeED25519,
		PublicKey: sysPubKeyData.PubKey,
		Signed:    mainEnvfileData.SignatureInfos.Signature,
		Data:      mainEnvfileData.EnvMapInfos,
	}); err != nil {
		return err
	}

	if err := v.ValidateExternalEnvMapData(mainEnvfileData.EnvMapInfos); err != nil {
		return err
	}

	if err := env.SetNewEnvMap[string, e.EnvData[[]byte]](env.MainEnvMapField, mainEnvfileData.EnvMapInfos); err != nil {
		return err
	}

	return nil
}

func includeInternalEnv() error {
	//TODO: tek seferlik sonrasında silinen imzalı file okuma işlemi yapılacak.
	//sistem setupları için yaml okuması //basit seviyedeki elle müdahale edilecek config yapıları için kullanılır.

	/*
		- systemWhiteListData.EnvInfos imzasının kontrol edilmesi için sistem pub key belirtilen path üzerinden okunur.
		Sonraki steplerde kullanılması için pubKey env eklenir.
	*/
	sysPubKeyData, err := getPubKey(env.MainPathEnvsPathKey, env.SystemPubKeyField)
	if err != nil {
		return err
	}

	//çağrılan system pub key pub key box kaydedilir sonraki steplerde kullanılması durumundan
	if err := env.SetNewPubKey(env.SystemKey, *sysPubKeyData); err != nil {
		return err
	}

	if err := incSysWhitelist(sysPubKeyData); err != nil {
		return err
	}

	if err := incMainEnv(sysPubKeyData); err != nil {
		return err
	}

	return nil
}

func includeExternalEnv(input e.IncludeExternalEnvInput) error {
	//external env sisteme ekleyebilmek için owner yetki ve yetkinin imzasını içeren oad(owner access data) almak gerekir.
	//sisteme tanımlanan whitelist içerisinde belirtilen owner access datasını getirmek için owner ait WhitelistData getirilir.

	//gönderilen verinin status durumu kontrol edilir.
	// if err := u.CheckDataStatusInfos(&e.CheckDataStatusInfosInput{
	// 	Status:      input.OwnerWhitelistKeyInfo.WhitelistKeyInfo.StatusInfos.Status,
	// 	ActiveAt:    input.OwnerWhitelistKeyInfo.WhitelistKeyInfo.StatusInfos.ActiveAt,
	// 	ExpiresAt:   input.OwnerWhitelistKeyInfo.WhitelistKeyInfo.StatusInfos.ExpiresAt,
	// 	Description: input.OwnerWhitelistKeyInfo.WhitelistKeyInfo.StatusInfos.Description,
	// }); err != nil {
	// 	return err
	// }

	//WhitelistData içerisindeki owner'ın belirttiği whitelist key ile varsa whitelist datası getirilir
	ownerWhitelistInfo, err := env.GetEnv[string, e.WhitelistData](env.WhitelistEnvMapField, input.OwnerWhitelistKeyInfo.WhitelistKeyInfo.OwnerKey)
	if err != nil {
		return err
	}

	//whitelist datasındaki belirtilen uri üzeirndeki owner imzası getirilir.
	ownerPubKey, err := getPubKey(env.SpecificPathKey, ownerWhitelistInfo.PathInfos.PubKeyDataURI)
	if err != nil {
		return err
	}

	//owner pub key ile ownerin beyan ettiği input.OwnerWhitelistKeyInfo imza kontrolü yapılır.
	if err := u.VerifySign(e.VerifySignInput[e.WhitelistKeyData]{
		SignType:  env.SignTypeED25519,
		PublicKey: ownerPubKey.PubKey,
		Signed:    input.OwnerWhitelistKeyInfo.SignatureInfos.Signature,
		Data:      input.OwnerWhitelistKeyInfo.WhitelistKeyInfo,
	}); err != nil {
		return err
	}

	//imza kontrolü tamamlanınca kullanılancak func göre files yetki durumu kontrol edilir.

	//access kontrolü gercekleştirilecek developer ait bilgilerle environment file engine hazırlanır.
	var ownerEnvFE *FileEngine[e.AccessData] = &FileEngine[e.AccessData]{
		Owner:      ownerWhitelistInfo.Owner,
		OADPathKey: env.SpecificPathKey,
		OADFields:  []string{ownerWhitelistInfo.PathInfos.AccessDataURI},
	}

	//external eklenecek env üzerindeki erişim bilgisi kontrol edilir.
	if err := ownerEnvFE.IFCheckAccessData(input.ReferenceTasks); err != nil {
		return err
	}

	//external eklenecek env üzerindeki erişimi uygunsa external envs sisteme eklenir.
	// external env bulunduğu base env path alınır
	externalBasePath, err := env.GetEnv(env.M)
	//external env yüklemesi belli bir sırayı takip ederek yüklenmek istenirse.
	currentKey := input.StartEnvMapField
	for currentKey != env.EndEnvMapField {
		//belirtilen key sahip env map mevcut mu kontrol edilir.
		envMapChainInfo, ok := input.EnvMapChainInfos[currentKey]
		if !ok {
			return env.GetFuncError(env.InvalidTaskAuthn, nil, currentKey)
		}

		//yüklemeler için genel env box oluşturulur
		var envBox map[string]e.EnvData = map[string]e.EnvData{}

		if err := ownerEnvFE.IFGet(e.GetInput[map[string]e.EnvData]{
			PathKey:    env.SystemExternalEnvMainPathKey,
			PathFields: []string{currentKey},
			Data:       &envBox,
		}); err != nil {
			return err
		}

		//external env yüklenebilmesi için tam kontrol sağlanır.
		if err := v.CheckEnvData(envBox); err != nil {
			return err
		}

		includeEnvMapfunc.IncludeFunc(envBox)

		currentKey = envMapChainInfo.NextEnvMap
	}

	return nil
}

func InitConfig() (map[string]string, error) {

	if err := includeInternalEnv(); err != nil {
		return nil, err
	}

	//access durumu kontrol edilen task-permissontype reference (refTaskPerr) ikilisi include edilir
	//access durumu kontrol edilecek task-permissontype reference ikilisi hazırlanır

	//yüklenme sırasına göre
	// includeEnvMapFuncs := map[string]func(input map[string]e.EnvData){
	// 	env.PathEnvField:      env.IncludePathEnvMap,
	// 	env.TaskEnvField:      env.IncludeTaskEnvMap,
	// 	env.RestEnvField:      env.IncludeRestEnvMap,
	// 	env.FuncEnvField:      env.IncludeFuncEnvMap,
	// 	env.FuncErrorEnvField: env.IncludeFuncErrorEnvMap,
	// }

	// includeEnvMapFuncInfos := map[string]string{
	// 	env.PathEnvsField:      env.TaskEnvsField,
	// 	env.TaskEnvsField:      env.RestEnvsField,
	// 	env.RestEnvsField:      env.FuncEnvsField,
	// 	env.FuncEnvsField:      env.FuncErrorEnvsField,
	// 	env.FuncErrorEnvsField: env.End,
	// }

	includeEnvMapFuncInfos := map[string]e.EnvMapChainData{
		env.PathEnvMapField:      e.EnvMapChainData{NextEnvMap: env.TaskEnvMapField, EnvKeyRefSlice: env.PathEnvKeyRefSlice},
		env.TaskEnvMapField:      e.EnvMapChainData{NextEnvMap: env.RestEnvMapField, EnvKeyRefSlice: env.TaskEnvKeyRefSlice},
		env.RestEnvMapField:      e.EnvMapChainData{NextEnvMap: env.FuncEnvMapField, EnvKeyRefSlice: env.RestEnvKeyRefSlice},
		env.FuncEnvMapField:      e.EnvMapChainData{NextEnvMap: env.FuncErrorEnvMapField, EnvKeyRefSlice: env.FuncEnvKeyRefSlice},
		env.FuncErrorEnvMapField: e.EnvMapChainData{NextEnvMap: env.EndEnvMapField, EnvKeyRefSlice: env.FuncErrorEnvKeyRefSlice},
	}

	// refTaskPerr := map[string]uint8{
	// 	env.PathEnvMapField:      env.RWPermType,
	// 	env.TaskEnvMapField:      env.RWSPermType,
	// 	env.RestEnvMapField:      env.RWBPermType,
	// 	env.FuncEnvMapField:      env.RWSBPermType,
	// 	env.FuncErrorEnvMapField: env.RWSPermType,
	// }

	if err := includeExternalEnv(e.IncludeExternalEnvInput{
		Owner:            owner.Value.(string), //dışardan request ile alınır(external)
		StartEnvMapField: env.PathEnvMapField,
		EnvMapChainInfos: includeEnvMapFuncInfos,
	}); err != nil {
		return nil, err
	}

	//bütün işlemler bitince
	includeEnvMapFuncInfos[env.MainPathEnvField] = e.IncludeEnvData{IncludeFunc: env.IncludeSystemMainPathEnvMap, NextIncludePath: env.PathEnvField}

	return includeEnvMapFuncInfos, nil
}

func ConfigFileWhatcher(configFiles ...string) error {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return env.GetFuncError(env.UnexpectedError, err)
	}
	defer watcher.Close()

	for _, path := range configFiles {
		err = watcher.Add(path)
		if err != nil {
			return env.GetFuncError(env.UnexpectedError, err)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return env.GetFuncError(env.UnexpectedError, err)
			}

			configHandleFileChange(event)
		case err, ok := <-watcher.Errors:
			if !ok {
				return env.GetFuncError(env.UnexpectedError, err)
			}
		}
	}
}

func configHandleFileChange(event fsnotify.Event) {

	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
		filepath.Base(event.Name)
		fmt.Printf("Dosya yazma işlemi algılandı: %s\n %v\n", event.Name, event.Op)
	case event.Op&fsnotify.Create == fsnotify.Create:
		fmt.Printf("name: %s\n", filepath.Base(event.Name))
		fmt.Printf("Dosya veya klasör oluşturuldu: %s\n", event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		fmt.Printf("name: %s\n", filepath.Base(event.Name))
		fmt.Printf("Dosya veya klasör silindi: %s\n", event.Name)
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		fmt.Printf("name: %s\n", filepath.Base(event.Name))
		fmt.Printf("Dosya veya klasör yeniden adlandırıldı: %s\n", event.Name)
	}
}
