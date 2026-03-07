package upgrade

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var releasesURL = "https://api.github.com/repos/nownow-labs/nownow/releases/latest"

func setReleasesURL(url string) {
	releasesURL = url
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func CheckLatest() (*Release, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(releasesURL)
	if err != nil {
		return nil, fmt.Errorf("checking latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing release: %w", err)
	}
	return &release, nil
}

func NormalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

func IsNewer(current, latest string) bool {
	return NormalizeVersion(current) != NormalizeVersion(latest)
}

func FindAsset(release *Release) (*Asset, error) {
	suffix := fmt.Sprintf("_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, a := range release.Assets {
		if strings.HasSuffix(a.Name, suffix) {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func Download(asset *Asset, destPath string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	return extractBinary(resp.Body, destPath)
}

func extractBinary(r io.Reader, destPath string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("decompressing: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading archive: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && strings.HasSuffix(hdr.Name, "nownow") {
			tmp := destPath + ".tmp"
			f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("creating temp file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				os.Remove(tmp)
				return fmt.Errorf("writing binary: %w", err)
			}
			f.Close()

			if err := os.Rename(tmp, destPath); err != nil {
				os.Remove(tmp)
				return fmt.Errorf("replacing binary: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("nownow binary not found in archive")
}
