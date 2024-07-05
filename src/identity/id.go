package identity

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"os"
	"path"

	"github.com/sag-enhanced/native-app/src/file"
	"github.com/sag-enhanced/native-app/src/helper"
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
	return base64.RawStdEncoding.EncodeToString(data)
}

func (identity *Identity) Seal(data []byte) ([]byte, error) {
	return helper.RSASeal(&identity.PrivateKey.PublicKey, data)
}

func (identity *Identity) Unseal(data []byte) ([]byte, error) {
	return helper.RSAUnseal(identity.PrivateKey, data)
}

func (identity *Identity) Save(fm *file.FileManager) error {
	data, err := x509.MarshalPKCS8PrivateKey(identity.PrivateKey)
	if err != nil {
		return err
	}
	idFile := path.Join(fm.Options.DataDirectory, "sage2.id")
	return fm.WriteFile(idFile, data, false)
}

func LoadIdentity(fm *file.FileManager) (*Identity, error) {
	idFileNew := path.Join(fm.Options.DataDirectory, "sage2.id")
	data, err := fm.ReadFile(idFileNew)
	if err != nil {
		// migration for old id file (pre b7)
		idFileOld := path.Join(fm.Options.DataDirectory, "sage.id")
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

	id := &Identity{PrivateKey: private}
	if err := id.Save(fm); err != nil {
		return nil, err
	}

	return id, nil
}
