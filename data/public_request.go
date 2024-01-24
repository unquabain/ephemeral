package data

import (
	"fmt"

	"github.com/google/uuid"
)

// PublicRequest contains the information to be packed in an envelope for the
// public part of the request, which may travel over public channels such as Slack.
type PublicRequest struct {
	ID          uuid.UUID
	Key         PublicKey
	Description string
}

// Encode creates a complementary key, encrypts the message, and returns
// a response object that can be decrypted with the corresponding private key
// that generated the public request.
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
