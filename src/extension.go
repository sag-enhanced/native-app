package app

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

func installExtensionFromGithub(name string, browser string, download string) error {
	resp, err := http.Get(download)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP request returned %d", resp.StatusCode)
	}

	dir := path.Join(getStoragePath(), "ext", browser, name)

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
			new_file, err := os.OpenFile(path.Join(dir, file.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			zip_file, err := file.Open()
			if err != nil {
				return err
			}

			_, err = io.Copy(new_file, zip_file)
			new_file.Close()
			zip_file.Close()
		}
	}

	return nil
}
