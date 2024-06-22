package bindings

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func (b *Bindings) EncryptionStatus() EncryptionStatus {
	return EncryptionStatus{
		Enabled: b.fm.Manifest != nil,
		Locked:  b.fm.Cipher == nil && b.fm.Manifest != nil,
	}
}

func (b *Bindings) EncryptionUnlock(password string) error {
	return b.fm.TryLoadKey(password)
}

func (b *Bindings) EncryptionLock() {
	b.fm.Cipher = nil
}

func (b *Bindings) EncryptionDisable() error {
	if b.fm.Manifest != nil && b.fm.Cipher == nil {
		return errors.New("need to decrypt files first")
	}
	os.Remove(path.Join(b.options.DataDirectory, "manifest.json"))
	errs := b.fm.UpdateFiles(true)
	b.fm.Cipher = nil
	b.fm.Manifest = nil
	if len(errs) > 0 {
		fmt.Println("Failed to update files", errs)
		return errs[0]
	}
	return nil
}

func (b *Bindings) EncryptionEnable(passwords []string) error {
	if b.fm.Manifest != nil && b.fm.Cipher == nil {
		return errors.New("need to decrypt files first")
	}
	err := b.fm.CreateKey(passwords)
	if err != nil {
		return err
	}
	errs := b.fm.UpdateFiles(false)
	if len(errs) > 0 {
		fmt.Println("Failed to update files", errs)
		return errs[0]
	}
	return nil
}

type EncryptionStatus struct {
	Enabled bool `json:"enabled"`
	Locked  bool `json:"locked"`
}
