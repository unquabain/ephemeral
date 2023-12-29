package data

import (
	"fmt"

	"github.com/google/uuid"
)

type PrivateRequest struct {
	ID          uuid.UUID
	Key         PrivateKey
	Description string
}

func (r PrivateRequest) Public() PublicRequest {
	return PublicRequest{
		ID:          r.ID,
		Key:         r.Key.Public(),
		Description: r.Description,
	}
}

func (r *PrivateRequest) Decode(response Response) ([]byte, error) {
	cipher, err := cipherFromKeys(r.Key, response.Key)
	if err != nil {
		return nil, fmt.Errorf(`unable to create cypher from secret: %w`, err)
	}
	plaintext, err := decrypt(response.Data, cipher)
	if err != nil {
		return nil, fmt.Errorf(`unable to decrypt data: %w`, err)
	}
	return plaintext, nil
}

func NewRequest(description string) (PrivateRequest, error) {
	var (
		r PrivateRequest
	)
	r.ID = uuid.New()
	if k, err := NewPrivateKey(RandomCurve()); err != nil {
		return r, fmt.Errorf(`unable to find appropriate curve: %w`, err)
	} else {
		r.Key = k
	}
	r.Description = description
	return r, nil
}
