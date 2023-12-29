package data

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

type PublicKey struct{ *ecdh.PublicKey }

func (pk PublicKey) MarshalBinary() ([]byte, error) {
	if pk, err := x509.MarshalPKIXPublicKey(pk.PublicKey); err != nil {
		return nil, fmt.Errorf(`could not create PKIX encoding of key: %w`, err)
	} else {
		return pk, nil
	}
}
func (pk PublicKey) MarshalText() ([]byte, error) {
	if pk, err := pk.MarshalBinary(); err != nil {
		return pk, err
	} else {
		b := make([]byte, base64.StdEncoding.EncodedLen(len(pk)))
		base64.StdEncoding.Encode(b, pk)
		return b, nil
	}
}

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

func (pk *PublicKey) UnmarshalText(text []byte) error {
	k := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	if n, err := base64.StdEncoding.Decode(k, text); err != nil {
		return fmt.Errorf(`could not decode binary of key: %w`, err)
	} else {
		k = k[:n]
	}
	return pk.UnmarshalBinary(k)
}

func (pk *PublicKey) Equal(other PublicKey) bool {
	return pk.PublicKey.Equal(other.PublicKey)
}

func (pk PublicKey) Curve() ecdh.Curve {
	return pk.PublicKey.Curve()
}
