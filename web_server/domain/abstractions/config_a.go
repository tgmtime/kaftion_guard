package abstractions

import e "web_server/domain/entities"

type IFileEngine[T any] interface {
	IFGet(input e.GetInput[T]) error
	IFExists(pathKey int, pathFields ...string) error
	IFGetRootFilePath(pathKey int, pathFields ...string) (string, error)
	IFAccessOperation(input e.WhitelistAccessData) error
}
