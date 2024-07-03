package file

import (
	"crypto/cipher"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/sag-enhanced/native-app/src/options"
)

type FileHeader byte

const (
	FileHeaderRaw            FileHeader = 0x0
	FileHeaderEncrypted      FileHeader = 0x1
	FileHeaderCompressed     FileHeader = 0x2
	FileHeaderEncryptedNoPad FileHeader = 0x3
)

type FileManager struct {
	Manifest *EncryptionManifest
	Cipher   *cipher.Block
	Options  *options.Options
}

func NewFileManager(options *options.Options) (*FileManager, error) {
	fm := &FileManager{
		Options: options,
	}
	manifestPath := path.Join(options.DataDirectory, "manifest.json")
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
		// only load from the backup file if the main file is missing or otherwise unreadable
		bkp := filename + ".bkp"
		if content, err = os.ReadFile(bkp); err != nil {
			return nil, err
		}
		if err = os.Rename(bkp, filename); err != nil {
			return nil, err
		}
	}
	return fm.unpack(content)
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
	// we write to .bkp first to avoid corrupting the main file
	// then delete the main file and rename the .bkp to the main file
	bkp := filename + ".bkp"
	if err := os.WriteFile(bkp, packed, 0644); err != nil {
		return err
	}
	if err := os.Remove(filename); err != nil {
		return err
	}
	if err := os.Rename(bkp, filename); err != nil {
		return err
	}
	return nil
}

func (fm *FileManager) UpdateFiles(ignoreCipher bool) []error {
	errors := []error{}
	fileNames := []string{}
	if files, err := os.ReadDir(path.Join(fm.Options.DataDirectory, "data")); err == nil {
		for _, file := range files {
			fileNames = append(fileNames, path.Join(fm.Options.DataDirectory, "data", file.Name()))
		}
	}
	if files, err := os.ReadDir(fm.Options.DataDirectory); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".id") {
				fileNames = append(fileNames, path.Join(fm.Options.DataDirectory, file.Name()))
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
	return path.Join(fm.Options.DataDirectory, "data", name+".dat")
}
