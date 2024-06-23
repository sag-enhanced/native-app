package browser

import (
	"os"
	"path"

	"github.com/sag-enhanced/native-app/src/options"
)

func getExtensionList(options *options.Options, browser string) ([]string, error) {
	ext := path.Join(options.DataDirectory, "ext", browser)
	files, err := os.ReadDir(ext)
	extensions := []string{}
	if err != nil {
		return extensions, nil
	}
	for _, file := range files {
		if file.IsDir() {
			extensions = append(extensions, path.Join(ext, file.Name()))
		}
	}
	return extensions, nil
}
