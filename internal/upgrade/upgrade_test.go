package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"v0.2.2", "0.2.2"},
		{"0.2.2", "0.2.2"},
		{"v1.0.0", "1.0.0"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := NormalizeVersion(tt.input); got != tt.want {
			t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
	}{
		// Same version
		{"0.2.2", "0.2.2", false},
		{"v0.2.2", "0.2.2", false},
		// Patch upgrade
		{"0.2.1", "0.2.2", true},
		// Minor upgrade
		{"0.2.9", "0.3.0", true},
		// Major upgrade
		{"1.9.9", "2.0.0", true},
		// Dev build (non-semver current) → any release is newer
		{"dev", "0.2.2", true},
		// Downgrade must NOT trigger
		{"0.3.0", "0.2.9", false},
		{"1.0.0", "0.9.9", false},
		// Non-semver latest → don't upgrade to it
		{"0.2.2", "beta", false},
		// Both non-semver
		{"dev", "nightly", false},
	}
	for _, tt := range tests {
		if got := IsNewer(tt.current, tt.latest); got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestFindAsset(t *testing.T) {
	release := &Release{
		TagName: "v0.3.0",
		Assets: []Asset{
			{Name: "nownow_darwin_amd64.tar.gz", BrowserDownloadURL: "https://example.com/amd64"},
			{Name: "nownow_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/arm64"},
			{Name: "nownow_0.3.0_checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	asset, err := FindAsset(release)
	if err != nil {
		t.Fatalf("FindAsset() error: %v", err)
	}
	if asset == nil {
		t.Fatal("FindAsset() returned nil")
	}
}

func TestFindAssetNoMatch(t *testing.T) {
	release := &Release{
		TagName: "v0.3.0",
		Assets: []Asset{
			{Name: "nownow_linux_amd64.tar.gz"},
		},
	}

	_, err := FindAsset(release)
	if err == nil {
		t.Fatal("FindAsset() expected error for missing platform, got nil")
	}
}

func TestCheckLatest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v0.9.9","assets":[{"name":"nownow_darwin_arm64.tar.gz","browser_download_url":"http://example.com/dl"}]}`))
	}))
	defer server.Close()

	origURL := releasesURL
	setReleasesURL(server.URL)
	defer setReleasesURL(origURL)

	release, err := CheckLatest()
	if err != nil {
		t.Fatalf("CheckLatest() error: %v", err)
	}
	if release.TagName != "v0.9.9" {
		t.Errorf("TagName = %q, want %q", release.TagName, "v0.9.9")
	}
}

func TestCheckLatestHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	origURL := releasesURL
	setReleasesURL(server.URL)
	defer setReleasesURL(origURL)

	_, err := CheckLatest()
	if err == nil {
		t.Fatal("CheckLatest() expected error on 403, got nil")
	}
}

func makeTarGz(t *testing.T, filename string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name: filename,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func TestExtractBinary(t *testing.T) {
	payload := []byte("fake-binary-content")
	archive := makeTarGz(t, "nownow", payload)

	dest := filepath.Join(t.TempDir(), "nownow")
	if err := extractBinary(bytes.NewReader(archive), dest); err != nil {
		t.Fatalf("extractBinary() error: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading dest: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("content mismatch: got %q, want %q", got, payload)
	}

	info, _ := os.Stat(dest)
	if info.Mode().Perm() != 0755 {
		t.Errorf("permissions = %o, want 0755", info.Mode().Perm())
	}
}

func TestExtractBinaryNested(t *testing.T) {
	payload := []byte("nested-binary")
	archive := makeTarGz(t, "nownow_darwin_arm64/nownow", payload)

	dest := filepath.Join(t.TempDir(), "nownow")
	if err := extractBinary(bytes.NewReader(archive), dest); err != nil {
		t.Fatalf("extractBinary() error: %v", err)
	}

	got, _ := os.ReadFile(dest)
	if !bytes.Equal(got, payload) {
		t.Errorf("content mismatch: got %q, want %q", got, payload)
	}
}

func TestExtractBinaryNotFound(t *testing.T) {
	archive := makeTarGz(t, "README.md", []byte("hello"))

	dest := filepath.Join(t.TempDir(), "nownow")
	err := extractBinary(bytes.NewReader(archive), dest)
	if err == nil {
		t.Fatal("extractBinary() expected error when binary not in archive")
	}
}

func TestDownload(t *testing.T) {
	payload := []byte("real-binary")
	archive := makeTarGz(t, "nownow", payload)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(archive)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "nownow")
	asset := &Asset{Name: "nownow_darwin_arm64.tar.gz", BrowserDownloadURL: server.URL}

	if err := Download(asset, dest); err != nil {
		t.Fatalf("Download() error: %v", err)
	}

	got, _ := os.ReadFile(dest)
	if !bytes.Equal(got, payload) {
		t.Errorf("content mismatch")
	}
}

func TestDownloadHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "nownow")
	asset := &Asset{Name: "test.tar.gz", BrowserDownloadURL: server.URL}

	err := Download(asset, dest)
	if err == nil {
		t.Fatal("Download() expected error on 404")
	}
}
