package image

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestIsRemoteURL(t *testing.T) {
	if !isRemoteURL("https://example.com/image.jpg") {
		t.Fatal("expected https URL to be remote")
	}
	if !isRemoteURL("http://example.com/image.jpg") {
		t.Fatal("expected http URL to be remote")
	}
	if isRemoteURL("file:///tmp/image.jpg") {
		t.Fatal("expected file URL to be non-remote")
	}
	if isRemoteURL("/tmp/image.jpg") {
		t.Fatal("expected local path to be non-remote")
	}
}

func TestFileURLPath(t *testing.T) {
	pathValue, ok, err := fileURLPath("file:///tmp/My%20Pic.png")
	if !ok {
		t.Fatal("expected file URL to be detected")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pathValue != filepath.FromSlash("/tmp/My Pic.png") {
		t.Fatalf("unexpected path: %s", pathValue)
	}
}

func TestExtractFileName(t *testing.T) {
	name := extractFileName("https://example.com/images/photo.jpg?size=large")
	if name != "photo.jpg" {
		t.Fatalf("unexpected name: %s", name)
	}

	name = extractFileName("https://example.com/images/")
	if name != "wallpaper" {
		t.Fatalf("unexpected fallback name: %s", name)
	}
}

func TestExtensionFromContentType(t *testing.T) {
	if extensionFromContentType(normalizeMediaType("image/jpeg; charset=utf-8")) != ".jpg" {
		t.Fatal("expected .jpg")
	}
	if extensionFromContentType(normalizeMediaType("image/png")) != ".png" {
		t.Fatal("expected .png")
	}
	if extensionFromContentType(normalizeMediaType("text/plain")) != "" {
		t.Fatal("expected empty extension")
	}
}

func TestProcessLocalNoMove(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "pic.png")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	proc := NewProcessor(nil, bytes.NewReader([]byte{0x01, 0x02, 0x03}))
	got, err := proc.Process(ctx, filePath, dir, true)
	if err != nil {
		t.Fatalf("process: %v", err)
	}

	absPath, _ := filepath.Abs(filePath)
	if got != absPath {
		t.Fatalf("expected %s, got %s", absPath, got)
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestProcessLocalMove(t *testing.T) {
	ctx := context.Background()
	sourceDir := t.TempDir()
	saveDir := t.TempDir()
	filePath := filepath.Join(sourceDir, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	proc := NewProcessor(nil, bytes.NewReader([]byte{0x10, 0x20, 0x30}))
	got, err := proc.Process(ctx, filePath, saveDir, false)
	if err != nil {
		t.Fatalf("process: %v", err)
	}

	if _, err := os.Stat(got); err != nil {
		t.Fatalf("expected moved file: %v", err)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("expected original file removed, got: %v", err)
	}

	if filepath.Dir(got) != saveDir {
		t.Fatalf("expected save dir %s, got %s", saveDir, filepath.Dir(got))
	}
	base := filepath.Base(got)
	if !strings.HasPrefix(base, "photo_") || !strings.HasSuffix(base, ".jpg") {
		t.Fatalf("unexpected file name: %s", base)
	}
}

func TestDownloadImage(t *testing.T) {
	ctx := context.Background()

	client := &http.Client{
		Transport: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("pngdata")),
				Header:     http.Header{"Content-Type": []string{"image/png"}},
			}, nil
		}),
	}

	proc := NewProcessor(client, bytes.NewReader([]byte{0x01, 0x02, 0x03}))
	saveDir := t.TempDir()
	url := "https://example.test/images/sample"

	got, err := proc.Process(ctx, url, saveDir, false)
	if err != nil {
		t.Fatalf("process: %v", err)
	}

	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "pngdata" {
		t.Fatalf("unexpected file contents: %s", string(data))
	}
	if !strings.HasSuffix(got, ".png") {
		t.Fatalf("expected .png extension, got %s", got)
	}
}

func TestDownloadRejectsNonImage(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{
		Transport: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("nope")),
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
			}, nil
		}),
	}

	proc := NewProcessor(client, bytes.NewReader([]byte{0x01, 0x02, 0x03}))
	saveDir := t.TempDir()

	if _, err := proc.Process(ctx, "https://example.test/file.txt", saveDir, false); err == nil {
		t.Fatal("expected error for non-image content type")
	}
}
