package utils

import (
	e "web_server/domain/entities"
	env "web_server/environments/processors"

	"github.com/fxamacker/cbor/v2"
)

func VerifySign[T any](input e.VerifySignInput[T]) error {

	encodedAuthnInfos, err := cbor.Marshal(input.Data)
	if err != nil {
		return env.GetFuncError(env.UnexpectedError, err)
	}

	switch input.SignType {
	case env.SignTypeED25519:
		if err := VerifySignED25519(e.VerifySignED25519Input{
			PublicKey: input.PublicKey,
			Signed:    input.Signed,
			Data:      encodedAuthnInfos,
		}); err != nil {
			return err
		}
	default:
		return env.GetFuncError(env.InvalidSignType, nil)
	}
	//owner tarafından oluşturulan imzası kontrol edilir.

	return nil
}

func CheckAuthn(ownerAuthn map[string]e.TaskData, referenceAuthn map[string]uint8) error {
	for taskKey, permissonType := range referenceAuthn {

		//belirtilen taskKey sahip mi kontrol edilir.
		taskInfo, ok := ownerAuthn[taskKey]
		if !ok {
			return env.GetFuncError(env.InvalidTaskAuthn, nil, taskKey)
		}

		if taskInfo.PermissionType&permissonType == 0 {
			return env.GetFuncError(env.InvalidTaskAuthn, nil, taskKey)
		}

		//belirtilen taskKey sahipse, status bilgisi kontrol edilir.
		if err := CheckDataStatusInfos(&e.CheckDataStatusInfosInput{
			Status:      taskInfo.StatusInfos.Status,
			ActiveAt:    taskInfo.StatusInfos.ActiveAt,
			ExpiresAt:   taskInfo.StatusInfos.ExpiresAt,
			Description: taskInfo.StatusInfos.Description,
		}); err != nil {
			return err
		}
	}
	return nil
}
