package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func AESSeal(key []byte, data []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipher, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, cipher.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	sealed := cipher.Seal(nil, iv, data, nil)

	result := make([]byte, len(iv)+len(sealed))
	copy(result, iv)
	copy(result[len(iv):], sealed)

	return result, nil
}

func AESUnseal(key []byte, data []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipher, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, err
	}
	if len(data) < cipher.NonceSize() {
		return nil, errors.New("data is too short")
	}
	iv := data[:cipher.NonceSize()]
	sealed := data[cipher.NonceSize():]

	return cipher.Open(nil, iv, sealed, nil)
}
