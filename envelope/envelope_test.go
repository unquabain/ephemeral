package envelope_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Unquabain/ephemeral/envelope"
	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
)

type LogHandleAdapter struct{ *testing.T }

func (lha LogHandleAdapter) HandleLog(entry *log.Entry) error {
	lha.T.Helper()
	lha.T.Log(entry.Message)
	return nil
}

func TestEnvelope(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetHandler(LogHandleAdapter{t})
	assert := assert.New(t)
	subject := envelope.Envelope{
		Name: `TEST ENVELOPE`,
		Prelude: `Most people have a furious itch to talk about themselves and are restrained
only by the disinclination of others to listen.  Reserve is an artificial
quality that is developed in most of us as the result of innumerable rebuffs.
		-- W.S. Maugham`,
		Data: []byte(`Ah, my friends, from the prison, they ask unto me,
"How good, how good does it feel to be free?"
And I answer them most mysteriously:
"Are birds free from the chains of the sky-way?"
		-- Bob Dylan`),
		Postlude: `When users see one GUI as beautiful,
other user interfaces become ugly.
When users see some programs as winners,
other programs become lossage.

Pointers and NULLs reference each other.
High level and assembler depend on each other.
Double and float cast to each other.
High-endian and low-endian define each other.
While and until follow each other.

Therefore the Guru 
programs without doing anything
and teaches without saying anything.
Warnings arise and he lets them come;
processes are swapped and he lets them go.
He has but doesn't possess,
acts but doesn't expect.
When his work is done, he deletes it.
That is why it lasts forever.`,
	}
	var recovered envelope.Envelope
	encoded, err := subject.MarshalText()
	assert.NoError(err)
	assert.NoError(recovered.UnmarshalText(encoded), string(encoded))
	assert.Equal(subject, recovered)

}

func TestEnvelopeDataReaderWriter(t *testing.T) {
	var (
		assert  = assert.New(t)
		subject envelope.Envelope
		data    = []byte(`Molecule, n.:
	The ultimate, indivisible unit of matter.  It is distinguished
	from the corpuscle, also the ultimate, indivisible unit of matter, by a
	closer resemblance to the atom, also the ultimate, indivisible unit of
	matter ... The ion differs from the molecule, the corpuscle and the
	atom in that it is an ion ...
	-- Ambrose Bierce, "The Devil's Dictionary"`,
		)
		reader = bytes.NewReader(data)
		writer = new(bytes.Buffer)
	)
	_, err := io.Copy(subject.DataWriter(), reader)
	assert.NoError(err)
	assert.Equal(data, subject.Data)

	_, err = io.Copy(writer, subject.DataReader())
	assert.NoError(err)
	assert.Equal(data, writer.Bytes())
}
