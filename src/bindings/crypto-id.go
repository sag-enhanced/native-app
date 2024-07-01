package bindings

import (
	"encoding/base64"
	"errors"

	"github.com/sag-enhanced/native-app/src/file"
	id "github.com/sag-enhanced/native-app/src/identity"
)

var identity *id.Identity

func (b *Bindings) Id() (string, error) {
	id, err := getIdentity(b.fm)
	if err != nil {
		return "", err
	}
	return id.Id(), nil
}

func (b *Bindings) Sign2(message string) (string, error) {
	identity, err := getIdentity(b.fm)
	if err != nil {
		return "", err
	}
	decoded, err := base64.RawStdEncoding.DecodeString(message)
	if err != nil {
		return "", err
	}
	// 0x0001 is header for signature requests
	if len(decoded) < 2 || decoded[0] != 0x00 || decoded[1] != 0x01 {
		return "", errors.New("Invalid message")
	}
	signature, err := identity.Sign(decoded)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(signature), nil
}

func (b *Bindings) Seal(data string) (string, error) {
	identity, err := getIdentity(b.fm)
	if err != nil {
		return "", err
	}
	plaintext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	sealed, err := identity.Seal(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (b *Bindings) Unseal(data string) (string, error) {
	identity, err := getIdentity(b.fm)
	if err != nil {
		return "", err
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	unsealed, err := identity.Unseal(decoded)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(unsealed), nil

}
func getIdentity(fm *file.FileManager) (*id.Identity, error) {
	if identity == nil {
		id, err := id.LoadIdentity(fm)
		if err != nil {
			return nil, err
		}
		identity = id
	}
	return identity, nil
}
