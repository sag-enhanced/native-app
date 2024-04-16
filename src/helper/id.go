package helper

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path"

	"github.com/sag-enhanced/native-app/src/file"
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

func (identity *Identity) Seal(data []byte) ([]byte, error) {
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

	sealed := cipher.Seal(nil, iv, file.Pad(data, aes.BlockSize), nil)
	sealedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &identity.private.PublicKey, key, nil)
	if err != nil {
		return nil, err
	}

	result := make([]byte, len(iv)+len(sealedKey)+len(sealed))
	copy(result, sealedKey)
	copy(result[len(sealedKey):], iv)
	copy(result[len(sealedKey)+len(iv):], sealed)

	return result, nil
}

func (identity *Identity) Unseal(data []byte) ([]byte, error) {
	sealedKeyLen := 512
	key, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, identity.private, data[:sealedKeyLen], nil)
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

	return file.Unpad(plain), nil
}

func LoadIdentity(fm *file.FileManager) (*Identity, error) {
	idFileNew := path.Join(file.GetStoragePath(), "sage2.id")
	data, err := fm.ReadFile(idFileNew)
	if err != nil {
		// migration for old id file (pre b7)
		idFileOld := path.Join(file.GetStoragePath(), "sage.id")
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
	err = fm.WriteFile(idFileNew, data, false)
	if err != nil {
		return nil, err
	}

	return &Identity{private: private}, nil
}
