package helper

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
	PrivateKey *rsa.PrivateKey
}

func (identity *Identity) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	return rsa.SignPSS(rand.Reader, identity.PrivateKey, crypto.SHA256, hash[:], nil)
}

func (identity *Identity) Id() string {
	data := x509.MarshalPKCS1PublicKey(&identity.PrivateKey.PublicKey)
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: data}))
}

func (identity *Identity) Seal(data []byte) ([]byte, error) {
	return RSASeal(&identity.PrivateKey.PublicKey, data)
}

func (identity *Identity) Unseal(data []byte) ([]byte, error) {
	return RSAUnseal(identity.PrivateKey, data)
}

type fileManager interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, dontEncrypt bool) error
}

func LoadIdentity(fm fileManager) (*Identity, error) {
	idFileNew := path.Join(GetStoragePath(), "sage2.id")
	data, err := fm.ReadFile(idFileNew)
	if err != nil {
		// migration for old id file (pre b7)
		idFileOld := path.Join(GetStoragePath(), "sage.id")
		data, err = os.ReadFile(idFileOld)
		if err == nil {
			err = fm.WriteFile(idFileNew, data, false)
			if err == nil {
				os.Remove(idFileOld)
			}
		}
	}
	if err == nil {
		private, err := x509.ParsePKCS8PrivateKey(data)
		if err != nil {
			return nil, err
		}
		if private, ok := private.(*rsa.PrivateKey); ok {
			return &Identity{PrivateKey: private}, nil
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
	err = fm.WriteFile(idFileNew, data, false)
	if err != nil {
		return nil, err
	}

	return &Identity{PrivateKey: private}, nil
}
