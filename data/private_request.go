package data

import (
	"fmt"

	"github.com/google/uuid"
)

// PrivateRequest contains the data necessary to make a full request.
type PrivateRequest struct {
	ID          uuid.UUID
	Key         PrivateKey
	Description string
}

// Public returns the corresponding PublicRequest object, which
// has the same data, except for it has the PublicKey that corresponds
// to this PrivateRequest's PrivateKey.
func (r PrivateRequest) Public() PublicRequest {
	return PublicRequest{
		ID:          r.ID,
		Key:         r.Key.Public(),
		Description: r.Description,
	}
}

// Decode extracts the PublicKey from the response and decrypts the
// response payload.
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

// NewRequest creates a new request with a random private key.
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
