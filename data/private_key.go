package data

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// PrivateKey is a wrapper around ecdh.PrivateKey decorating it with functions
// we'll use in this package.
type PrivateKey struct{ *ecdh.PrivateKey }

// NewPrivateKey creates a new key given a valid Elliptic Curve.
func NewPrivateKey(c ecdh.Curve) (PrivateKey, error) {
	if k, err := c.GenerateKey(rand.Reader); err != nil {
		return PrivateKey{}, fmt.Errorf(`unable to generate key from curve: %w`, err)
	} else {
		return PrivateKey{k}, nil
	}
}

// Public returns the corresponding PublicKey decorator for this PrivateKey.
func (pk *PrivateKey) Public() PublicKey {
	return PublicKey{pk.PrivateKey.PublicKey()}
}

// MarshalBinary implements encoding.BinaryMarshaler, and is used
// to pack it in an envelope.
func (pk PrivateKey) MarshalBinary() ([]byte, error) {
	if pk, err := x509.MarshalPKCS8PrivateKey(pk.PrivateKey); err != nil {
		return pk, fmt.Errorf(`unable to marshal private key: %w`, err)
	} else {
		return pk, nil
	}
}

// MarshalText implements encoding.TextMarshaler, and is used to make YAML
// representations.
func (pk PrivateKey) MarshalText() ([]byte, error) {
	if pk, err := pk.MarshalBinary(); err != nil {
		return nil, fmt.Errorf(`could not creat PKCS8 encoding of key: %w`, err)
	} else {
		b := make([]byte, base64.StdEncoding.EncodedLen(len(pk)))
		base64.StdEncoding.Encode(b, pk)
		return b, nil
	}
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler, and is used to
// extract the PrivateKey from an envelope.
func (pk *PrivateKey) UnmarshalBinary(data []byte) error {
	if k, err := x509.ParsePKCS8PrivateKey(data); err != nil {
		return fmt.Errorf(`could not understand PrivateKey as i509 PKCS8 Public Key: %w`, err)
	} else if k, ok := k.(*ecdsa.PrivateKey); !ok {
		return fmt.Errorf(`could not understand PrivateKey as i509 ECDSA Public Key: %w`, err)
	} else if k, err := k.ECDH(); err != nil {
		return fmt.Errorf(`could not understand PrivateKey as i509 ECDH Public Key: %w`, err)
	} else {
		pk.PrivateKey = k
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler, and is used to read the value from YAML.
func (pk *PrivateKey) UnmarshalText(text []byte) error {
	k := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	if n, err := base64.StdEncoding.Decode(k, text); err != nil {
		return fmt.Errorf(`could not decode binary of key: %w`, err)
	} else {
		k = k[:n]
	}
	return pk.UnmarshalBinary(k)
}

// Secret takes a PublicKey and returns a []byte, the value of which can only be determined
// in two ways: by having this private key and that public key, or by having the corresponding
// public and private keys, respectively. That is, with Alice's private key and Bob's public key,
// you can find the same value as with Bob's private key and Alice's public key. However, there
// is no way to determine this value given both public keys.
func (pk *PrivateKey) Secret(pub PublicKey) ([]byte, error) {
	return pk.PrivateKey.ECDH(pub.PublicKey)
}

// Equal wraps the underlying Equal method. It is useful for tests.
func (pk *PrivateKey) Equal(other PrivateKey) bool {
	return pk.PrivateKey.Equal(other.PrivateKey)
}
