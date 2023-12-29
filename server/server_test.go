//go:build test_internal
// +build test_internal

package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/Unquabain/ephemeral/envelope"
	"github.com/tj/assert"
)

func makeBodyInto[T any](val T) func(any) error {
	return func(target any) error {
		tp := target.(*T)
		*tp = val
		return nil
	}
}

func extractJSON(r response, target any) error {
	recorder := httptest.NewRecorder()
	if err := r.Respond(recorder); err != nil {
		return err
	}
	result := recorder.Result()
	defer result.Body.Close()
	return json.NewDecoder(result.Body).Decode(target)
}

func extractEnvelope(r response, target *envelope.Envelope) error {
	recorder := httptest.NewRecorder()
	if err := r.Respond(recorder); err != nil {
		return err
	}
	result := recorder.Result()
	defer result.Body.Close()
	_, err := target.ReadFrom(result.Body)
	return err
}

func extractBytes(r response) ([]byte, error) {
	recorder := httptest.NewRecorder()
	if err := r.Respond(recorder); err != nil {
		return nil, err
	}
	result := recorder.Result()
	defer result.Body.Close()
	data := new(bytes.Buffer)
	_, err := io.Copy(data, result.Body)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func TestServer(t *testing.T) {
	var (
		description = `Absolute:  Independent, irresponsible.  An absolute monarchy is one in which
the sovereign does as he pleases so long as he pleases the assassins.  Not
many absolute monarchies are left, most of them having been replaced by
limited monarchies, where the soverign's power for evil (and for good) is
greatly curtailed, and by republics, which are governed by chance.
-- Ambrose Bierce`
		secret = `/*
 * [...] Note that 120 sec is defined in the protocol as the maximum
 * possible RTT.  I guess we'll have to use something other than TCP
 * to talk to the University of Mars.
 * PAWS allows us longer timeouts and large windows, so once implemented
 * ftp to mars will work nicely.
 */
(from /usr/src/linux/net/inet/tcp.c, concerning RTT [retransmission timeout])`
		requestRequest  = struct{ Description string }{Description: description}
		requestResponse struct {
			PrivateRequest envelope.Envelope
			PublicRequest  envelope.Envelope
		}
		respondRequest = struct {
			PublicRequest envelope.Envelope
			Data          string
		}{Data: secret}
		respondResponse envelope.Envelope
		receiveRequest  struct {
			PrivateRequest envelope.Envelope
			Data           envelope.Envelope
		}
		assert = assert.New(t)
	)

	r, werr := request(makeBodyInto(requestRequest))
	assert.Nil(werr)
	assert.NoError(extractJSON(r, &requestResponse))

	respondRequest.PublicRequest = requestResponse.PublicRequest
	r, werr = respond(makeBodyInto(respondRequest))
	assert.Nil(werr)
	assert.NoError(extractEnvelope(r, &respondResponse))

	receiveRequest.PrivateRequest = requestResponse.PrivateRequest
	receiveRequest.Data = respondResponse
	r, werr = receive(makeBodyInto(receiveRequest))
	assert.Nil(werr)
	text, err := extractBytes(r)
	assert.NoError(err)

	assert.Equal([]byte(secret), text)
}
