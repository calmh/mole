package ticket

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"errors"
)

const (
	keySize  = 32
	ivSize   = 16
	hashSize = 20
)

var (
	key []byte
	iv  []byte
)

func init() {
	initKeyAndIV()
}

func initKeyAndIV() {
	key = make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	iv = make([]byte, ivSize)
	_, err = rand.Read(iv)
	if err != nil {
		panic(err)
	}
}

func encrypt(blob []byte) {
	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	s := cipher.NewCFBEncrypter(c, iv)
	s.XORKeyStream(blob, blob)
}

func decrypt(blob []byte) {
	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	s := cipher.NewCFBDecrypter(c, iv)
	s.XORKeyStream(blob, blob)
}

func hashAndFrame(blob []byte) []byte {
	h := sha1.New()
	_, err := h.Write(blob)
	if err != nil {
		panic(err)
	}
	hash := h.Sum(nil)

	var b bytes.Buffer

	_, err = b.Write(hash)
	if err != nil {
		panic(err)
	}

	_, err = b.Write(blob)
	if err != nil {
		panic(err)
	}

	return b.Bytes()
}

func hashAndEncrypt(blob []byte) []byte {
	frame := hashAndFrame(blob)
	encrypt(frame)
	return frame
}

func unframe(blob []byte) ([]byte, []byte, error) {
	b := bytes.NewBuffer(blob)
	var err error

	if hashSize > uint32(b.Len()) {
		return nil, nil, errors.New("corrupt packet")
	}
	hash := make([]byte, hashSize)
	_, err = b.Read(hash)
	if err != nil {
		return nil, nil, err
	}

	msg := b.Bytes()
	return msg, hash, nil
}

func decryptAndHash(blob []byte) ([]byte, error) {
	decrypt(blob)

	msg, hash, err := unframe(blob)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	_, err = h.Write(msg)
	if err != nil {
		panic(err)
	}
	comp := h.Sum(nil)

	if c := bytes.Compare(hash, comp); c != 0 {
		return nil, errors.New("hash failure")
	}

	return msg, nil
}
