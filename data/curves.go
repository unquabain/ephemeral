package data

import (
	"crypto/ecdh"
	"crypto/rand"

	"github.com/apex/log"
)

// Curve is an enum for selecting an elliptic curve.
type Curve uint8

const (
	// P256 is the P256 elliptic curve
	P256 Curve = iota

	// P384 is the P256 elliptic curve
	P384

	// P521 is the P256 elliptic curve
	P521

	// InvalidCurve is a constant indicating an error choosing a random curve.
	InvalidCurve
)

// RandomCurve selects a supported, secure elliptic curve at random.
func RandomCurve() ecdh.Curve {
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		log.WithError(err).Fatal(`unable to read a random byte`)
	}
	c := Curve(b[0]) % InvalidCurve
	switch c {
	case P256:
		return ecdh.P256()
	case P384:
		return ecdh.P384()
	case P521:
		return ecdh.P521()
	}
	log.Fatal(`read unreadable random byte`)
	return nil
}
