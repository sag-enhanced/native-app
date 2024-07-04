package bindings

import (
	"crypto/x509"
	"encoding/base64"

	"github.com/sag-enhanced/native-app/src/helper"
)

func (b *Bindings) SealWithPublicKey(data string, publicKey string) (string, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", err
	}
	pk, err := x509.ParsePKCS1PublicKey(decoded)
	if err != nil {
		return "", err
	}

	plaintext, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	sealed, err := helper.RSASeal(pk, plaintext)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func (b *Bindings) SealWithKey(data string, key string) (string, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	plaintext, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	sealed, err := helper.AESSeal(decoded, plaintext)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func (b *Bindings) UnsealWithKey(data string, key string) (string, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	plaintext, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	unsealed, err := helper.AESUnseal(decoded, plaintext)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(unsealed), nil
}
