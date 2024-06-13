package bindings

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/sag-enhanced/native-app/src/file"
)

var start = time.Now().UnixMilli()

func (b *Bindings) Build() uint32 {
	return b.options.Build
}
func (b *Bindings) Start() int64 {
	return start
}
func (b *Bindings) Info() (map[string]any, error) {
	id, _ := machineid.ID()
	exe := os.Args[0]
	var exeHash string

	if content, err := os.ReadFile(exe); err == nil {
		digest := sha256.Sum256(content)
		exeHash = hex.EncodeToString(digest[:])
	}

	return map[string]any{
		"build":    b.options.Build,
		"path":     file.GetStoragePath(),
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"id":       id,
		"port":     b.options.LoopbackPort,
		"args":     os.Args,
		"exe":      exe,
		"exe_hash": exeHash,
		"url":      b.currentUrl.String(),
	}, nil
}

func (b *Bindings) Quit() {
	b.ui.Quit()
}
