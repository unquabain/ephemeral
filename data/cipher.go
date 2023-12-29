package data

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/apex/log"
)

const keySize = 32

func cipherFromKeys(private PrivateKey, public PublicKey) (cipher.Block, error) {
	secret, err := private.Secret(public)
	if err != nil {
		return nil, fmt.Errorf(`unable to create shared secret: %w`, err)
	}
	key := make([]byte, keySize)
	for len(secret) < keySize {
		secret = append(secret, secret...)
	}
	for i, b := range secret {
		key[i%keySize] = key[i%keySize] ^ b
	}
	log.Debugf(`%X`, key)
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf(`unable to create cypher from secret: %w`, err)
	}
	return cipher, nil
}

func encrypt(data []byte, block cipher.Block) ([]byte, error) {
	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf(`could not create random IV: %w`, err)
	}
	out := new(bytes.Buffer)
	if _, err := out.Write(iv); err != nil {
		return nil, fmt.Errorf(`could not write IV to writer %w`, err)
	}
	stream := cipher.NewOFB(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: out}
	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf(`could not encrypt data to writer: %w`, err)
	}
	return out.Bytes(), nil
}

func decrypt(data []byte, block cipher.Block) ([]byte, error) {
	blockSize := block.BlockSize()
	iv := data[:blockSize]
	data = data[blockSize:]
	stream := cipher.NewOFB(block, iv)
	reader := &cipher.StreamReader{S: stream, R: bytes.NewReader(data)}
	buff := new(bytes.Buffer)
	if _, err := io.Copy(buff, reader); err != nil {
		return nil, fmt.Errorf(`could not encrypt data from reader: %w`, err)
	}
	return buff.Bytes(), nil
}
