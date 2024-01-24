package data

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// PublicKey decorates ecdh.PublicKey with methods we will use in this package.
type PublicKey struct{ *ecdh.PublicKey }

// MarshalBinary implements encoding.BinaryMarshaler, and is used to insert
// the PublicKey in an envelope.
func (pk PublicKey) MarshalBinary() ([]byte, error) {
	if pk, err := x509.MarshalPKIXPublicKey(pk.PublicKey); err != nil {
		return nil, fmt.Errorf(`could not create PKIX encoding of key: %w`, err)
	} else {
		return pk, nil
	}
}

// MarshalText implements encoding.TextMarshaler, and is used to create YAML
func (pk PublicKey) MarshalText() ([]byte, error) {
	if pk, err := pk.MarshalBinary(); err != nil {
		return pk, err
	} else {
		b := make([]byte, base64.StdEncoding.EncodedLen(len(pk)))
		base64.StdEncoding.Encode(b, pk)
		return b, nil
	}
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler, and is used to extract
// the PublicKey from an envelope.
func (pk *PublicKey) UnmarshalBinary(data []byte) error {
	if k, err := x509.ParsePKIXPublicKey(data); err != nil {
		return fmt.Errorf(`could not understand PublicKey as i509 PKIX Public Key: %w`, err)
	} else if k, ok := k.(*ecdsa.PublicKey); !ok {
		return fmt.Errorf(`could not understand PublicKey as i509 ECDSA Public Key: %w`, err)
	} else if k, err := k.ECDH(); err != nil {
		return fmt.Errorf(`could not understand PublicKey as i509 ECDH Public Key: %w`, err)
	} else {
		pk.PublicKey = k
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler, and is used to retrieve the value from YAML
func (pk *PublicKey) UnmarshalText(text []byte) error {
	k := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	if n, err := base64.StdEncoding.Decode(k, text); err != nil {
		return fmt.Errorf(`could not decode binary of key: %w`, err)
	} else {
		k = k[:n]
	}
	return pk.UnmarshalBinary(k)
}

// Equal wraps the method of the underlying type, and is used in tests.
func (pk *PublicKey) Equal(other PublicKey) bool {
	return pk.PublicKey.Equal(other.PublicKey)
}

// Curve returns the curve used in the key pair, which is needed to make a complementary
// key pair.
func (pk PublicKey) Curve() ecdh.Curve {
	return pk.PublicKey.Curve()
}
