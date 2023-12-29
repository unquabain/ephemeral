package data

import (
	"crypto/ecdh"
	"crypto/rand"

	"github.com/apex/log"
)

type Curve uint8

const (
	P256 Curve = iota
	P384
	P521
	InvalidCurve
)

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
