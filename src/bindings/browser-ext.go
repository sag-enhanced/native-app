package bindings

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
	if err := validatePublicDownloadURL(download); err != nil {
		return errors.New("invalid download URL")
	}
	if path.Clean(name) != name || strings.Contains(name, ",") {
		return errors.New("invalid extension name")
	}

	return installExtensionFromURL(name, browser, download, b.options)
}

func validatePublicDownloadURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || strings.Contains(u.Path, "..") {
		return errors.New("invalid download URL")
	}

	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") || host == "home.arpa" || strings.HasSuffix(host, ".home.arpa") {
		return errors.New("invalid download URL")
	}

	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return errors.New("invalid download URL")
		}
		return nil
	}

	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return errors.New("invalid download URL")
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return errors.New("invalid download URL")
		}
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	return ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsUnspecified() && !ip.IsMulticast()
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

func installExtensionFromURL(name string, browser string, download string, options *options.Options) error {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			if err := validatePublicDownloadURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
	}
	resp, err := client.Get(download)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP request returned %d", resp.StatusCode)
	}

	dir := path.Join(options.DataDirectory, "ext", browser, name)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		// ZIP entry names always use slash separators. Reject paths that could
		// escape the extension directory on either Unix or Windows.
		entryName := file.Name
		cleanName := path.Clean(entryName)
		firstComponent := strings.SplitN(cleanName, "/", 2)[0]
		if strings.Contains(entryName, "\\") || strings.Contains(firstComponent, ":") || path.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, "../") {
			return fmt.Errorf("invalid ZIP entry path %q", entryName)
		}
		if cleanName == "." {
			continue
		}

		target := filepath.Join(dir, filepath.FromSlash(cleanName))
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		mode := file.Mode().Perm()
		if mode == 0 {
			mode = 0644
		}
		newFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		zipFile, err := file.Open()
		if err != nil {
			newFile.Close()
			return err
		}
		_, copyErr := io.Copy(newFile, zipFile)
		closeErr := newFile.Close()
		zipCloseErr := zipFile.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if zipCloseErr != nil {
			return zipCloseErr
		}
	}

	return nil
}
