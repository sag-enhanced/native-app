package helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

func Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func Unpad(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}

func RSASeal(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
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

	// we dont need to actually pad here because its GCM, but its too late to change it now
	sealed := cipher.Seal(nil, iv, Pad(data, aes.BlockSize), nil)
	sealedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, key, nil)
	if err != nil {
		return nil, err
	}

	result := make([]byte, len(iv)+len(sealedKey)+len(sealed))
	copy(result, sealedKey)
	copy(result[len(sealedKey):], iv)
	copy(result[len(sealedKey)+len(iv):], sealed)

	return result, nil
}

func RSAUnseal(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	sealedKeyLen := 512
	key, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data[:sealedKeyLen], nil)
	if err != nil {
		return nil, err
	}
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipher, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, err
	}

	iv := data[sealedKeyLen : sealedKeyLen+cipher.NonceSize()]
	sealed := data[sealedKeyLen+cipher.NonceSize():]

	plain, err := cipher.Open(nil, iv, sealed, nil)
	if err != nil {
		return nil, err
	}

	return Unpad(plain), nil
}
