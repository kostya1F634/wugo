package image

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultSuffixLength = 6
	defaultTimeout      = 30 * time.Second
)

type Processor struct {
	client *http.Client
	rand   io.Reader
}

func NewProcessor(client *http.Client, randReader io.Reader) *Processor {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	if randReader == nil {
		randReader = rand.Reader
	}
	return &Processor{client: client, rand: randReader}
}

func (p *Processor) Process(ctx context.Context, input, saveDir string, noMove bool) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", errors.New("empty input")
	}

	if filePath, ok, err := fileURLPath(input); ok {
		if err != nil {
			return "", err
		}
		input = filePath
	}

	if isRemoteURL(input) {
		return p.download(ctx, input, saveDir)
	}

	return p.handleLocal(input, saveDir, noMove)
}

func (p *Processor) handleLocal(filePath, saveDir string, noMove bool) (string, error) {
	absPath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return "", fmt.Errorf("absolute path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", absPath)
	}

	if noMove {
		return absPath, nil
	}

	fileName := filepath.Base(absPath)
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if base == "" {
		base = "wallpaper"
	}
	ext := filepath.Ext(fileName)
	suffix := p.uniqueSuffix(defaultSuffixLength)
	newFileName := fmt.Sprintf("%s_%s%s", base, suffix, ext)
	destPath := filepath.Join(saveDir, newFileName)

	if err := os.Rename(absPath, destPath); err != nil {
		if err := copyFile(absPath, destPath); err != nil {
			return "", fmt.Errorf("copy file: %w", err)
		}
		if err := os.Remove(absPath); err != nil {
			return "", fmt.Errorf("remove original: %w", err)
		}
	}

	return filepath.Abs(destPath)
}

func (p *Processor) download(ctx context.Context, imageURL, saveDir string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := p.httpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	mediaType := normalizeMediaType(resp.Header.Get("Content-Type"))
	if mediaType != "" && !strings.HasPrefix(mediaType, "image/") {
		return "", fmt.Errorf("unexpected content type: %s", mediaType)
	}

	name := extractFileName(imageURL)
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if base == "" {
		base = "wallpaper"
	}

	ext := filepath.Ext(name)
	if ext == "" {
		ext = extensionFromContentType(mediaType)
	}

	suffix := p.uniqueSuffix(defaultSuffixLength)
	finalName := fmt.Sprintf("%s_%s%s", base, suffix, ext)
	localPath := filepath.Join(saveDir, finalName)

	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}

	copyErr := copyToFile(out, resp.Body)
	if copyErr != nil {
		_ = os.Remove(localPath)
		return "", copyErr
	}

	return filepath.Abs(localPath)
}

func isRemoteURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func fileURLPath(input string) (string, bool, error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", false, nil
	}
	if u.Scheme != "file" {
		return "", false, nil
	}
	if u.Host != "" && u.Host != "localhost" {
		return "", true, fmt.Errorf("unsupported file URL host: %s", u.Host)
	}
	if u.Path == "" {
		return "", true, errors.New("empty file URL path")
	}
	pathValue, err := url.PathUnescape(u.Path)
	if err != nil {
		return "", true, fmt.Errorf("decode file URL: %w", err)
	}
	return filepath.FromSlash(pathValue), true, nil
}

func extractFileName(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "wallpaper"
	}
	if strings.HasSuffix(u.Path, "/") {
		return "wallpaper"
	}
	name := path.Base(u.Path)
	if name == "/" || name == "." || name == "" {
		return "wallpaper"
	}
	return name
}

func normalizeMediaType(contentType string) string {
	if contentType == "" {
		return ""
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err == nil {
		return strings.ToLower(mediaType)
	}
	parts := strings.Split(contentType, ";")
	return strings.ToLower(strings.TrimSpace(parts[0]))
}

func extensionFromContentType(mediaType string) string {
	switch mediaType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	case "image/svg+xml":
		return ".svg"
	default:
		return ""
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func copyToFile(dst *os.File, src io.Reader) error {
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return err
	}
	return dst.Close()
}

func (p *Processor) uniqueSuffix(n int) string {
	return randomHex(n, p.randReader())
}

func randomHex(n int, r io.Reader) string {
	if n <= 0 {
		return ""
	}

	byteLen := (n + 1) / 2
	buf := make([]byte, byteLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return strings.Repeat("x", n)
	}

	hexStr := hex.EncodeToString(buf)
	return hexStr[:n]
}

func (p *Processor) randReader() io.Reader {
	if p.rand != nil {
		return p.rand
	}
	return rand.Reader
}

func (p *Processor) httpClient() *http.Client {
	if p.client != nil {
		return p.client
	}
	return &http.Client{Timeout: defaultTimeout}
}
