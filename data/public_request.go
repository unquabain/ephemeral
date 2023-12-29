package data

import (
	"fmt"

	"github.com/google/uuid"
)

type PublicRequest struct {
	ID          uuid.UUID
	Key         PublicKey
	Description string
}

func (r PublicRequest) Encode(data []byte) (Response, error) {
	privateKey, err := NewPrivateKey(r.Key.Curve())
	if err != nil {
		return Response{}, fmt.Errorf(`unable to create private key: %w`, err)
	}
	cipher, err := cipherFromKeys(privateKey, r.Key)
	if err != nil {
		return Response{}, fmt.Errorf(`unable to create cypher from secret: %w`, err)
	}
	ciphertext, err := encrypt(data, cipher)
	if err != nil {
		return Response{}, fmt.Errorf(`unable to encrypt data: %w`, err)
	}
	return Response{
		ID:   r.ID,
		Data: ciphertext,
		Key:  privateKey.Public(),
	}, nil
}
