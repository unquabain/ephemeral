package envelope

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"strings"
)

type Envelope struct {
	Name     string
	Prelude  string
	Data     []byte
	Postlude string
}

const wrapLength = 64

type wrapper struct {
	io.WriteCloser
	ll int
}

func zip(data []byte) ([]byte, error) {
	in := bytes.NewReader(data)
	out := new(bytes.Buffer)
	if w, err := zlib.NewWriterLevel(out, zlib.BestCompression); err != nil {
		return nil, fmt.Errorf(`could not create compression writer: %w`, err)
	} else if _, err := io.Copy(w, in); err != nil {
		return nil, fmt.Errorf(`could not compress data: %w`, err)
	} else if err := w.Close(); err != nil {
		return nil, fmt.Errorf(`could not finalize compressed data: %w`, err)
	}
	return out.Bytes(), nil
}

func unzip(data []byte) ([]byte, error) {
	in := bytes.NewReader(data)
	out := new(bytes.Buffer)
	if reader, err := zlib.NewReader(in); err != nil {
		return nil, fmt.Errorf(`could not create new zlib reader: %w`, err)
	} else if _, err := io.Copy(out, reader); err != nil {
		return nil, fmt.Errorf(`could not decompress data: %w`, err)
	}
	return out.Bytes(), nil
}

func encode(data []byte) ([]byte, error) {
	in := bytes.NewReader(data)
	out := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, out)
	if _, err := io.Copy(enc, in); err != nil {
		return nil, fmt.Errorf(`could not encode data: %w`, err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf(`could not finalize encoding: %w`, err)
	}
	return out.Bytes(), nil
}

func decode(data []byte) ([]byte, error) {
	in := bytes.NewReader(data)
	out := new(bytes.Buffer)
	if _, err := io.Copy(out, base64.NewDecoder(base64.StdEncoding, in)); err != nil {
		return nil, fmt.Errorf(`could not decode data: %w`, err)
	}
	return out.Bytes(), nil
}

func wrap(data []byte) ([]byte, error) {
	out := new(bytes.Buffer)
	for len(data) >= wrapLength {
		if _, err := out.Write(data[:wrapLength]); err != nil {
			return nil, fmt.Errorf(`could not wrap data: %w`, err)
		}
		if _, err := out.WriteRune('\n'); err != nil {
			return nil, fmt.Errorf(`could not wrap data: %w`, err)
		}
		data = data[wrapLength:]
	}
	if len(data) > 0 {
		if _, err := out.Write(data); err != nil {
			return nil, fmt.Errorf(`could not wrap data: %w`, err)
		}
		if _, err := out.WriteRune('\n'); err != nil {
			return nil, fmt.Errorf(`could not wrap data: %w`, err)
		}
	}

	return out.Bytes(), nil
}

func (e Envelope) MarshalText() ([]byte, error) {
	buff := new(bytes.Buffer)
	fmt.Fprintln(buff, e.Prelude)
	fmt.Fprintf(buff, "----- BEGIN %s -----\n", strings.ToUpper(e.Name))

	if data, err := zip(e.Data); err != nil {
		return nil, fmt.Errorf(`could not zip data: %w`, err)
	} else if data, err := encode(data); err != nil {
		return nil, fmt.Errorf(`could not encode data: %w`, err)
	} else if data, err := wrap(data); err != nil {
		return nil, fmt.Errorf(`could not wrap data: %w`, err)
	} else if _, err := buff.Write(data); err != nil {
		return nil, fmt.Errorf(`could not write data: %w`, err)
	}

	fmt.Fprintf(buff, "----- END %s -----\n", strings.ToUpper(e.Name))
	fmt.Fprintln(buff, e.Postlude)
	return buff.Bytes(), nil
}

func (e *Envelope) UnmarshalText(data []byte) error {
	parts := strings.Split(string(data), `-----`)
	if l := len(parts); l != 5 {
		return fmt.Errorf(`not enough dash-delimited parts: expected 5, found %d`, l)
	}
	e.Prelude = strings.TrimSpace(parts[0])
	e.Name = strings.TrimSpace(strings.TrimPrefix(parts[1], ` BEGIN `))

	if data, err := decode([]byte(parts[2])); err != nil {
		return fmt.Errorf(`unable to decode data: %w`, err)
	} else if data, err := unzip(data); err != nil {
		return fmt.Errorf(`unable to unzip data: %w`, err)
	} else {
		e.Data = data
	}

	e.Postlude = strings.TrimSpace(parts[4])
	return nil
}

type errorReader struct{ error }

func (e errorReader) Read(_ []byte) (int, error) {
	return 0, e.error
}

func (e Envelope) Reader() io.Reader {
	if data, err := e.MarshalText(); err != nil {
		return errorReader{err}
	} else {
		return bytes.NewReader(data)
	}
}

func (e *Envelope) ReadFrom(r io.Reader) (int64, error) {
	buff := new(bytes.Buffer)
	n, err := io.Copy(buff, r)
	if err != nil {
		return int64(n), err
	}
	return int64(n), e.UnmarshalText(buff.Bytes())
}

func (e Envelope) DataReader() io.Reader {
	return bytes.NewReader(e.Data)
}

type buffWriter struct{ *Envelope }

func (b buffWriter) Write(data []byte) (int, error) {
	buff := bytes.NewBuffer(b.Envelope.Data)
	n, err := buff.Write(data)
	b.Envelope.Data = buff.Bytes()
	return n, err
}

func (e *Envelope) DataWriter() io.Writer {
	return buffWriter{e}
}

func (e *Envelope) Stuff(content any) error {
	return gob.NewEncoder(e.DataWriter()).Encode(content)
}

func (e *Envelope) Open(target any) error {
	return gob.NewDecoder(e.DataReader()).Decode(target)
}
