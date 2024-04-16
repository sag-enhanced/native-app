package bindings

import (
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
	id, err := machineid.ID()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"build": b.options.Build,
		"path":  file.GetStoragePath(),
		"os":    runtime.GOOS,
		"arch":  runtime.GOARCH,
		"id":    id,
		"port":  b.options.LoopbackPort,
	}, nil
}

func (b *Bindings) Quit() {
	b.ui.Quit()
}
