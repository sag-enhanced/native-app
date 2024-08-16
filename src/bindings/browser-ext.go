package bindings

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/sag-enhanced/native-app/src/options"
)

func (b *Bindings) Ext(browser string) (*map[string]string, error) {
	dir := path.Join(b.options.DataDirectory, "ext", browser)
	extensions := map[string]string{}
	files, err := os.ReadDir(dir)
	// probably directory doesnt exist, so no extensions
	if err != nil {
		return &extensions, nil
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		manifest, err := os.ReadFile(path.Join(dir, file.Name(), "manifest.json"))
		if err != nil {
			continue
		}
		var parsedManifest Manifest
		err = json.Unmarshal(manifest, &parsedManifest)
		if err != nil {
			continue
		}

		extensions[file.Name()] = parsedManifest.Version
	}

	return &extensions, nil
}

func (b *Bindings) ExtInstall(name string, browser string, download string) error {
	if !strings.HasPrefix(download, "https://github.com/") || strings.Contains(download, "..") {
		return errors.New("invalid download URL")
	}
	if path.Clean(name) != name || strings.Contains(name, ",") {
		return errors.New("invalid extension name")
	}

	return installExtensionFromGithub(name, browser, download, b.options)
}

func (b *Bindings) ExtGetManifest(name string, browser string) (string, error) {
	if path.Clean(name) != name || strings.Contains(name, ",") {
		return "", errors.New("invalid extension name")
	}

	manifest := path.Join(b.options.DataDirectory, "ext", browser, name, "manifest.json")

	data, err := os.ReadFile(manifest)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (b *Bindings) ExtSetManifest(name string, browser string, manifest string) error {
	if path.Clean(name) != name || strings.Contains(name, ",") {
		return errors.New("invalid extension name")
	}

	manifestPath := path.Join(b.options.DataDirectory, "ext", browser, name, "manifest.json")
	return os.WriteFile(manifestPath, []byte(manifest), 0644)
}

func (b *Bindings) ExtUninstall(name string, browser string) error {
	if path.Clean(name) != name || strings.Contains(name, ",") {
		return errors.New("invalid extension name")
	}

	dir := path.Join(b.options.DataDirectory, "ext", browser, name)
	return os.RemoveAll(dir)
}

type Manifest struct {
	Version string `json:"version"`
}

func installExtensionFromGithub(name string, browser string, download string, options *options.Options) error {
	resp, err := http.Get(download)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP request returned %d", resp.StatusCode)
	}

	dir := path.Join(options.DataDirectory, "ext", browser, name)

	os.MkdirAll(dir, 0755)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(body), resp.ContentLength)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			os.MkdirAll(path.Join(dir, file.Name), 0755)
		} else {
			newFile, err := os.OpenFile(path.Join(dir, file.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			zipFile, err := file.Open()
			if err != nil {
				return err
			}

			_, err = io.Copy(newFile, zipFile)
			newFile.Close()
			zipFile.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
