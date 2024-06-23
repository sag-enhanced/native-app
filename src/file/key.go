package file

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/sag-enhanced/native-app/src/helper"
)

func (fm *FileManager) TryLoadKey(password string) error {
	if fm.Manifest == nil {
		return errors.New("no manifest")
	}

	salt, err := hex.DecodeString(fm.Manifest.Salt)
	if err != nil {
		return err
	}
	key := helper.DeriveKey(password, salt)

	control := sha256.Sum256(key)
	encoded := hex.EncodeToString(control[:])
	for _, k := range fm.Manifest.Keys {
		// we dont need to use a constant time comparison here because the hash is already public
		if k.Hash == encoded {
			secret, err := hex.DecodeString(k.Secret)
			if err != nil {
				return err
			}
			decryptCipher, err := aes.NewCipher(key)
			if err != nil {
				return err
			}
			decrypted := make([]byte, len(secret))
			decryptCipher.Decrypt(decrypted[:16], secret[:16])
			decryptCipher.Decrypt(decrypted[16:], secret[16:])

			cipher, err := aes.NewCipher(decrypted)
			if err != nil {
				return err
			}
			fm.Cipher = &cipher
			return nil
		}
	}

	return errors.New("invalid password")
}

func (fm *FileManager) CreateKey(passwords []string) error {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return err
	}
	keys := make([]EncryptionKey, len(passwords))
	for i, password := range passwords {
		key := helper.DeriveKey(password, salt)

		cipher, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		encryptedMasterKey := make([]byte, 32)
		cipher.Encrypt(encryptedMasterKey[:16], masterKey[:16])
		cipher.Encrypt(encryptedMasterKey[16:], masterKey[16:])

		hashedKey := sha256.Sum256(key)
		keys[i] = EncryptionKey{
			Hash:   hex.EncodeToString(hashedKey[:]),
			Secret: hex.EncodeToString(encryptedMasterKey),
		}
	}
	manifest := EncryptionManifest{
		Version: 1,
		Salt:    hex.EncodeToString(salt),
		Keys:    keys,
	}

	manifestPath := path.Join(fm.Options.DataDirectory, "manifest.json")
	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	err = os.WriteFile(manifestPath, data, 0644)
	if err != nil {
		return err
	}
	cipher, err := aes.NewCipher(masterKey)
	if err != nil {
		return err
	}
	fm.Manifest = &manifest
	fm.Cipher = &cipher
	return nil
}
