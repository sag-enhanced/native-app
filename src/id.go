package app

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path"
)

type Identity struct {
	private *rsa.PrivateKey
}

func (identity *Identity) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	return rsa.SignPSS(rand.Reader, identity.private, crypto.SHA256, hash[:], nil)
}

func (identity *Identity) Id() string {
	data := x509.MarshalPKCS1PublicKey(&identity.private.PublicKey)
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: data}))
}

func loadIdentity() (*Identity, error) {
	idFile := path.Join(getStoragePath(), "sage.id")
	data, err := os.ReadFile(idFile)
	if err == nil {
		private, err := x509.ParsePKCS8PrivateKey(data)
		if err != nil {
			return nil, err
		}
		if private, ok := private.(*rsa.PrivateKey); ok {
			return &Identity{private: private}, nil
		}
		return nil, errors.New("invalid private key")
	}
	private, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	data, err = x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(idFile, data, 0644)
	if err != nil {
		return nil, err
	}

	return &Identity{private: private}, nil
}
