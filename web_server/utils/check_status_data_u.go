package utils

import (
	"time"
	e "web_server/domain/entities"
	env "web_server/environments/processors"
)

// new data status bilgilerin stadartları kontrol edilir.
func CheckNewDataStatusInfos(input *e.CheckDataStatusInfosInput) error {

	if !input.Status {
		return env.GetFuncError(env.InvalidNewDataStatus, nil)
	}
	//zamansal kontrol için kullanılır.
	checkUnixTime := time.Now().Unix()
	//expiresAt durumu 0 değil ve şimdiki zamandan küçük olma durumu kontrol edilir.
	if checkUnixTime > input.ExpiresAt && input.ExpiresAt != 0 {
		return env.GetFuncError(env.InvalidNewDataExpiresAt, nil)
	}

	if checkUnixTime > input.ActiveAt {
		return env.GetFuncError(env.InvalidNewDataActiveAt, nil)
	}

	if input.Description == "" {
		return env.GetFuncError(env.InvalidNewDataStatusDescription, nil)
	}

	return nil
}


// data status bilgileri kontrol edilir.
func CheckDataStatusInfos(input *e.CheckDataStatusInfosInput) error {

	if !input.Status {
		return env.GetFuncError(env.InactiveDataStatus, nil)
	}
	//zamansal kontrol için kullanılır.
	checkUnixTime := time.Now().Unix()
	//expiresAt durumu 0 değil ve şimdiki zamandan küçük olma durumu kontrol edilir.
	if checkUnixTime > input.ExpiresAt && input.ExpiresAt != 0 {
		return env.GetFuncError(env.DataExpiresAt, nil)
	}

	if checkUnixTime < input.ActiveAt {
		return env.GetFuncError(env.DataActiveAt, nil)
	}

	if input.Description == "" {
		return env.GetFuncError(env.RequiredDataStatusDescription, nil)
	}

	return nil
}