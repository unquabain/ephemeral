package data_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/Unquabain/ephemeral/data"
	"github.com/stretchr/testify/assert"
)

func TestPrivateRequestMarshal(t *testing.T) {
	assert := assert.New(t)
	request, err := data.NewRequest(``)
	assert.NoError(err)
	privateEncoded := new(bytes.Buffer)
	assert.NoError(gob.NewEncoder(privateEncoded).Encode(request))
	var recovered data.PrivateRequest
	assert.NoError(gob.NewDecoder(privateEncoded).Decode(&recovered))
	assert.True(request.Key.Equal(recovered.Key))
	assert.Equal(request.ID, recovered.ID)
}

func TestPublicRequestMarshal(t *testing.T) {
	assert := assert.New(t)
	request, err := data.NewRequest(``)
	assert.NoError(err)
	public := request.Public()
	publicEncoded := new(bytes.Buffer)
	assert.NoError(gob.NewEncoder(publicEncoded).Encode(public))
	var recovered data.PublicRequest
	assert.NoError(gob.NewDecoder(publicEncoded).Decode(&recovered))
	assert.True(public.Key.Equal(recovered.Key))
	assert.Equal(public.ID, recovered.ID)
}

func TestEncodeDecode(t *testing.T) {
	assert := assert.New(t)
	message := []byte(`HOW TO PROVE IT, PART 1

proof by example:
	The author gives only the case n = 2 and suggests that it 
	contains most of the ideas of the general proof.

proof by intimidation:
	'Trivial'.

proof by vigorous handwaving:
	Works well in a classroom or seminar setting.
`)

	private, err := data.NewRequest(``)
	assert.NoError(err)
	public := private.Public()

	encrypted, err := public.Encode(message)

	assert.NoError(err)
	assert.NotEqual(encrypted.Data, message)

	decrypted, err := private.Decode(encrypted)
	assert.NoError(err)

	assert.Equal(decrypted, message)

}
