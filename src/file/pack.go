package file

import (
	"bytes"
	"compress/flate"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"github.com/sag-enhanced/native-app/src/helper"
)

func (fm *FileManager) unpack(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty file (corrupted)")
	}
	header := FileHeader(data[0])
	if header == FileHeaderRaw {
		return data[1:], nil
	} else if header == FileHeaderEncrypted || header == FileHeaderEncryptedNoPad {
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

		if header == FileHeaderEncryptedNoPad {
			return fm.unpack(content)
		}
		// GCM does not need padding, but due to an historical false judgement, we still padded it
		// FileHeaderEncrypted is legacy and will not be created anymore
		return fm.unpack(helper.Unpad(content))
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

func (fm *FileManager) pack(data []byte, ignoreCipher bool) ([]byte, error) {
	data = append([]byte{byte(FileHeaderRaw)}, data...)
	data, err := fm.tryCompress(data)
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
		encrypted := aesCipher.Seal(nil, nonce, data, nil)
		data = make([]byte, 1+len(nonce)+len(encrypted))
		data[0] = byte(FileHeaderEncryptedNoPad)
		copy(data[1:], nonce)
		copy(data[1+len(nonce):], encrypted)
	}
	data, err = fm.tryCompress(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (fm *FileManager) tryCompress(data []byte) ([]byte, error) {
	// compression was disabled
	if fm.Options.NoCompress {
		return data, nil
	}
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
