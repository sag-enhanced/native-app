package bindings

import (
	"encoding/base64"
	"errors"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/sag-enhanced/native-app/src/helper"
)

func (b *Bindings) FsReadFile(filename string) (string, error) {
	path, err := validateFilename(filename)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	if utf8.Valid(content) {
		return string(content), nil
	}
	return "data:;base64," + base64.StdEncoding.EncodeToString(content), nil
}

func (b *Bindings) FsWriteFile(filename string, content string) error {
	path, err := validateFilename(filename)
	if err != nil {
		return err
	}

	if strings.HasPrefix(content, "data:") {
		decoded, err := base64.StdEncoding.DecodeString(strings.SplitN(content, ",", 2)[1])
		if err != nil {
			return err
		}
		content = string(decoded)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func (b *Bindings) FsDeleteFile(filename string) error {
	path, err := validateFilename(filename)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

func (b *Bindings) FsListFiles(dirname string) ([]string, error) {
	path, err := validateFilename(dirname)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}

func (b *Bindings) FsMkdir(dirname string) error {
	path, err := validateFilename(dirname)
	if err != nil {
		return err
	}

	return os.MkdirAll(path, 0755)
}

func validateFilename(filename string) (string, error) {
	cleaned := path.Clean(strings.ReplaceAll(filename, "\\", "/"))
	if cleaned != filename {
		return "", errors.New("Invalid filename")
	}

	realName := path.Clean(path.Join(helper.GetStoragePath(), "files", filename))
	if !strings.HasPrefix(realName, helper.GetStoragePath()) {
		return "", errors.New("Invalid filename")
	}
	return realName, nil
}
