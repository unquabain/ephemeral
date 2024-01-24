package data

import "github.com/google/uuid"

// Response represents encrypted data that can be shared over public channels.
type Response struct {
	ID   uuid.UUID
	Key  PublicKey
	Data []byte
}
