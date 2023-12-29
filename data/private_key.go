package data

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

type PrivateKey struct{ *ecdh.PrivateKey }

func NewPrivateKey(c ecdh.Curve) (PrivateKey, error) {
	if k, err := c.GenerateKey(rand.Reader); err != nil {
		return PrivateKey{}, fmt.Errorf(`unable to generate key from curve: %w`, err)
	} else {
		return PrivateKey{k}, nil
	}
}

func (pk *PrivateKey) Public() PublicKey {
	return PublicKey{pk.PrivateKey.PublicKey()}
}

func (pk PrivateKey) MarshalBinary() ([]byte, error) {
	if pk, err := x509.MarshalPKCS8PrivateKey(pk.PrivateKey); err != nil {
		return pk, fmt.Errorf(`unable to marshal private key: %w`, err)
	} else {
		return pk, nil
	}
}
func (pk PrivateKey) MarshalText() ([]byte, error) {
	if pk, err := pk.MarshalBinary(); err != nil {
		return nil, fmt.Errorf(`could not creat PKCS8 encoding of key: %w`, err)
	} else {
		b := make([]byte, base64.StdEncoding.EncodedLen(len(pk)))
		base64.StdEncoding.Encode(b, pk)
		return b, nil
	}
}

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

func (pk *PrivateKey) UnmarshalText(text []byte) error {
	k := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	if n, err := base64.StdEncoding.Decode(k, text); err != nil {
		return fmt.Errorf(`could not decode binary of key: %w`, err)
	} else {
		k = k[:n]
	}
	return pk.UnmarshalBinary(k)
}

func (pk *PrivateKey) Secret(pub PublicKey) ([]byte, error) {
	return pk.PrivateKey.ECDH(pub.PublicKey)
}

func (pk *PrivateKey) Equal(other PrivateKey) bool {
	return pk.PrivateKey.Equal(other.PrivateKey)
}
