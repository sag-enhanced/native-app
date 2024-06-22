package bindings

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"path"
)

func (b *Bindings) Get(key string) (string, error) {
	filenameNew := b.fm.GetFilename(key)
	data, err := b.fm.ReadFile(filenameNew)
	if err == nil {
		return string(data), nil
	}
	fmt.Println("Failed to read new file", filenameNew, err)

	// fallback to old file (pre b7)
	filenameOld := path.Join(b.options.DataDirectory, key+".dat")
	data, err = os.ReadFile(filenameOld)
	if err != nil {
		return "", err
	}
	reader := flate.NewReader(bytes.NewReader(data))
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	err = b.fm.WriteFile(filenameNew, decompressed, false)
	if err == nil {
		os.Remove(filenameOld)
	} else {
		fmt.Println("Failed to write decompressed data to new file", filenameNew)
	}
	return string(decompressed), nil
}

func (b *Bindings) Set(key string, value string) error {
	filename := b.fm.GetFilename(key)
	return b.fm.WriteFile(filename, []byte(value), false)
}
