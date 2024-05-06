package file

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/argon2"
)

type FileHeader byte

const (
	FileHeaderRaw        FileHeader = 0x0
	FileHeaderEncrypted  FileHeader = 0x1
	FileHeaderCompressed FileHeader = 0x2
)

type FileManager struct {
	Manifest *EncryptionManifest
	Cipher   *cipher.Block
}

func NewFileManager() (*FileManager, error) {
	fm := &FileManager{}
	manifestPath := path.Join(GetStoragePath(), "manifest.json")
	manifestContent, err := os.ReadFile(manifestPath)
	if err == nil {
		var manifest EncryptionManifest
		if err := json.Unmarshal(manifestContent, &manifest); err != nil {
			return nil, err
		}
		fm.Manifest = &manifest
		if manifest.Version != 1 {
			return nil, errors.New("unsupported manifest version")
		}
	}
	return fm, nil
}

type EncryptionManifest struct {
	Version int32           `json:"version"`
	Keys    []EncryptionKey `json:"keys"`
	Salt    string          `json:"salt"`
}

type EncryptionKey struct {
	Hash   string `json:"hash"`
	Secret string `json:"secret"`
}

func (fm *FileManager) ReadFile(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return fm.unpack(content)
}

func (fm *FileManager) unpack(data []byte) ([]byte, error) {
	header := FileHeader(data[0])
	if header == FileHeaderRaw {
		return data[1:], nil
	} else if header == FileHeaderEncrypted {
		if fm.Cipher == nil {
			return nil, errors.New("encrypted")
		}
		aesCipher, err := cipher.NewGCM(*fm.Cipher)
		if err != nil {
			return nil, err
		}
		content, err := aesCipher.Open(nil, data[1:1+aesCipher.NonceSize()], data[1+aesCipher.NonceSize():], nil)
		if err != nil {
			return nil, err
		}

		return fm.unpack(Unpad(content))
	} else if header == FileHeaderCompressed {
		reader := flate.NewReader(bytes.NewReader(data[1:]))
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return fm.unpack(decompressed)
	}
	return nil, errors.New("unknown header")
}

func (fm *FileManager) WriteFile(filename string, data []byte, ignoreCipher bool) error {
	packed, err := fm.pack(data, ignoreCipher)
	if err != nil {
		return err
	}
	parent := path.Dir(filename)
	if _, err := os.Stat(parent); os.IsNotExist(err) {
		if err := os.MkdirAll(parent, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(filename, packed, 0644)
}

func (fm *FileManager) UpdateFiles(ignoreCipher bool) []error {
	errors := []error{}
	fileNames := []string{}
	if files, err := os.ReadDir(path.Join(GetStoragePath(), "data")); err == nil {
		for _, file := range files {
			fileNames = append(fileNames, path.Join(GetStoragePath(), "data", file.Name()))
		}
	}
	if files, err := os.ReadDir(GetStoragePath()); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".id") {
				fileNames = append(fileNames, path.Join(GetStoragePath(), file.Name()))
			}
		}
	}
	for _, filename := range fileNames {
		data, err := fm.ReadFile(filename)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		err = fm.WriteFile(filename, data, ignoreCipher)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (fm *FileManager) GetFilename(name string) string {
	return path.Join(GetStoragePath(), "data", name+".dat")
}

func (fm *FileManager) DeriveKey(password string, salt []byte) []byte {
	memory := uint32(64 * 1024)
	time := uint32(3)
	threads := uint8(4)
	bytes := uint32(32)

	return argon2.IDKey([]byte(password), salt, time, memory, threads, bytes)
}

func (fm *FileManager) TryLoadKey(password string) error {
	if fm.Manifest == nil {
		return errors.New("no manifest")
	}

	salt, err := hex.DecodeString(fm.Manifest.Salt)
	if err != nil {
		return err
	}
	key := fm.DeriveKey(password, salt)

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
		key := fm.DeriveKey(password, salt)

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

	manifestPath := path.Join(GetStoragePath(), "manifest.json")
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

func (fm *FileManager) pack(data []byte, ignoreCipher bool) ([]byte, error) {
	data = append([]byte{byte(FileHeaderRaw)}, data...)
	data, err := tryCompress(data)
	if err != nil {
		return nil, err
	}
	if fm.Cipher != nil && !ignoreCipher {
		aesCipher, err := cipher.NewGCM(*fm.Cipher)
		if err != nil {
			return nil, err
		}
		nonce := make([]byte, aesCipher.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}
		encrypted := aesCipher.Seal(nil, nonce, Pad(data, aes.BlockSize), nil)
		data = make([]byte, 1+len(nonce)+len(encrypted))
		data[0] = byte(FileHeaderEncrypted)
		copy(data[1:], nonce)
		copy(data[1+len(nonce):], encrypted)
	}
	data, err = tryCompress(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func tryCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{byte(FileHeaderCompressed)})
	writer, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = writer.Write(data)
	if err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	if buf.Len() >= len(data) {
		return data, nil
	}
	return buf.Bytes(), nil
}

func Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func Unpad(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}
