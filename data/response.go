package data

import "github.com/google/uuid"

type Response struct {
	ID   uuid.UUID
	Key  PublicKey
	Data []byte
}
