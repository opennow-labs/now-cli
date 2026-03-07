package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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

// IsNewer returns true only if latest is strictly greater than current.
// Uses numeric semver comparison; falls back to true for non-numeric versions
// (e.g. "dev") to allow upgrading from dev builds.
func IsNewer(current, latest string) bool {
	c := NormalizeVersion(current)
	l := NormalizeVersion(latest)
	if c == l {
		return false
	}
	cp := parseVersion(c)
	lp := parseVersion(l)
	if lp == nil {
		// Latest is non-semver, don't upgrade to it
		return false
	}
	if cp == nil {
		// Current is non-semver (e.g. "dev"), any valid release is newer
		return true
	}
	for i := 0; i < 3; i++ {
		if lp[i] > cp[i] {
			return true
		}
		if lp[i] < cp[i] {
			return false
		}
	}
	return false
}

// parseVersion splits "1.2.3" into [1, 2, 3]. Returns nil if not valid semver.
func parseVersion(v string) []int {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

func FindAsset(release *Release) (*Asset, error) {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	suffix := fmt.Sprintf("_%s_%s%s", runtime.GOOS, runtime.GOARCH, ext)
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

	if strings.HasSuffix(asset.Name, ".zip") {
		return extractBinaryFromZip(resp.Body, destPath)
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

func extractBinaryFromZip(r io.Reader, destPath string) error {
	// zip.Reader needs io.ReaderAt, so buffer the whole response
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading zip: %w", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}

	binaryName := "nownow.exe"
	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if name != binaryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("opening zip entry: %w", err)
		}
		defer rc.Close()

		tmp := destPath + ".tmp"
		out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			os.Remove(tmp)
			return fmt.Errorf("writing binary: %w", err)
		}
		out.Close()

		if err := os.Rename(tmp, destPath); err != nil {
			os.Remove(tmp)
			return fmt.Errorf("replacing binary: %w", err)
		}
		return nil
	}
	return fmt.Errorf("nownow.exe not found in zip archive")
}
