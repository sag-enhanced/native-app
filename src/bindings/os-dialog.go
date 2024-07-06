package bindings

import (
	"encoding/base64"
	"os"

	"github.com/gen2brain/beeep"
	"github.com/sqweek/dialog"
)

func (b *Bindings) Save(filename string, data string) error {
	path, err := dialog.File().Title("Save file").SetStartFile(filename).Filter("All files", "*").Save()
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(data), 0644)
}

func (b *Bindings) Save2(filename string, data string) error {
  path, err := dialog.File().Title("Save file").SetStartFile(filename).Filter("All files", "*").Save()
  if err != nil {
    return err
  }

  decoded, err := base64.RawStdEncoding.DecodeString(data)
  if err != nil {
    return err
  }
  return os.WriteFile(path, decoded, 0644)
}

func (b *Bindings) Read(filterText string, filter string) (string, error) {
	path, err := dialog.File().Title("Open file").Filter(filterText, filter).Load()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (b *Bindings) Alert(message string) {
	dialog.Message(message).Title("Alert").Info()
}

func (b *Bindings) Notify(title string, message string, alert bool) error {
	err := beeep.Notify(title, message, "")
	if err != nil {
		return err
	}
	if alert {
		return beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	}
	return nil
}
