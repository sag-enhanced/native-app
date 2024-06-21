package bindings

import (
	"encoding/base64"
	"errors"

	"github.com/sag-enhanced/native-app/src/file"
	"github.com/sag-enhanced/native-app/src/helper"
)

var identity *helper.Identity

func (b *Bindings) Id() (string, error) {
	id, err := getIdentity(b.fm)
	if err != nil {
		return "", err
	}
	return id.Id(), nil
}

// TODO: remove in b10 (login will be disabled in 2025)
func (b *Bindings) Sign(message string) ([]byte, error) {
	id, err := getIdentity(b.fm)
	if err != nil {
		return nil, err
	}
	return id.Sign([]byte(message))
}

func (b *Bindings) Sign2(message string) ([]byte, error) {
	id, err := getIdentity(b.fm)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.RawStdEncoding.DecodeString(message)
	if err != nil {
		return nil, err
	}
	// 0x0001 is header for signature requests
	if len(decoded) < 2 || decoded[0] != 0x00 || decoded[1] != 0x01 {
		return nil, errors.New("Invalid message")
	}
	return id.Sign(decoded)
}

func (b *Bindings) Seal(data string) (string, error) {
	id, err := getIdentity(b.fm)
	if err != nil {
		return "", err
	}
	plaintext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	sealed, err := id.Seal(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (b *Bindings) Unseal(data string) (string, error) {
	id, err := getIdentity(b.fm)
	if err != nil {
		return "", err
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	unsealed, err := id.Unseal(decoded)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(unsealed), nil

}
func getIdentity(fm *file.FileManager) (*helper.Identity, error) {
	if identity == nil {
		id, err := helper.LoadIdentity(fm)
		if err != nil {
			return nil, err
		}
		identity = id
	}
	return identity, nil
}
