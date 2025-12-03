package entities

type CheckDataStatusInfosInput struct {
	Status      bool
	ActiveAt    int64
	ExpiresAt   int64
	Description string
}
