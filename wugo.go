package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
)

func main() {
	dir := flag.String("d", "", "Directory to save/move image")
	noMove := flag.Bool("nm", false, "Do not move local file, use it from current location")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: wallpaper [-d dir] [-nm] <image-url-or-path>")
		return
	}

	input := flag.Arg(0)

	var saveDir string
	if *dir != "" {
		saveDir = *dir
	} else {
		usr, _ := user.Current()
		saveDir = filepath.Join(usr.HomeDir, "wallpapers")
	}

	os.MkdirAll(saveDir, 0o755)

	localPath, err := processImage(input, saveDir, *noMove)
	if err != nil {
		fmt.Println("Failed to process image:", err)
		return
	}

	if err := updateDesktop(localPath); err != nil {
		fmt.Println("Failed to set desktop wallpaper:", err)
	}

	if err := updateLockscreen(localPath); err != nil {
		fmt.Println("Failed to set lockscreen wallpaper:", err)
	}

	fmt.Println("Wallpaper set successfully:", localPath)
}

func processImage(input, saveDir string, noMove bool) (string, error) {
	if isURL(input) {
		return downloadImage(input, saveDir)
	}

	input = filepath.Clean(input)
	return handleLocalFile(input, saveDir, noMove)
}

func isURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func handleLocalFile(filePath, saveDir string, noMove bool) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", absPath)
	}

	if noMove {
		return absPath, nil
	}

	fileName := filepath.Base(absPath)
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	ext := filepath.Ext(fileName)
	suffix := uniqueSuffix(6)
	newFileName := fmt.Sprintf("%s_%s%s", base, suffix, ext)
	destPath := filepath.Join(saveDir, newFileName)

	err = os.Rename(absPath, destPath)
	if err != nil {
		err = copyFile(absPath, destPath)
		if err != nil {
			return "", fmt.Errorf("failed to copy file: %w", err)
		}
		os.Remove(absPath)
	}

	return destPath, nil
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

func downloadImage(imageURL, saveDir string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := mimeToExt(contentType)
	name := extractFileName(imageURL)
	base := strings.TrimSuffix(name, filepath.Ext(name))

	if filepath.Ext(name) == "" && ext != "" {
		name = base + "." + ext
	}

	suffix := uniqueSuffix(6)
	finalName := fmt.Sprintf("%s_%s%s", base, suffix, filepath.Ext(name))
	localPath := filepath.Join(saveDir, finalName)

	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

func extractFileName(urlStr string) string {
	parts := strings.Split(strings.TrimRight(urlStr, "/"), "/")
	name := parts[len(parts)-1]
	if name == "" {
		return "wallpaper"
	}
	return name
}

func mimeToExt(mime string) string {
	switch {
	case strings.Contains(mime, "jpeg"):
		return "jpg"
	case strings.Contains(mime, "png"):
		return "png"
	case strings.Contains(mime, "webp"):
		return "webp"
	case strings.Contains(mime, "gif"):
		return "gif"
	case strings.Contains(mime, "bmp"):
		return "bmp"
	case strings.Contains(mime, "svg"):
		return "svg"
	default:
		return ""
	}
}

func uniqueSuffix(n int) string {
	b := make([]byte, n/2)
	_, err := rand.Read(b)
	if err != nil {
		return "xxxxxx"
	}
	return hex.EncodeToString(b)
}

func updateDesktop(image string) error {
	script := fmt.Sprintf(`var d = desktops(); for (let i in d) {
	d[i].wallpaperPlugin = "org.kde.image";
	d[i].currentConfigGroup = ["Wallpaper", "org.kde.image", "General"];
	d[i].writeConfig("Image", "file://%s");
}`, image)

	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	obj := conn.Object("org.kde.plasmashell", "/PlasmaShell")
	return obj.Call("org.kde.PlasmaShell.evaluateScript", 0, script).Err
}

func updateLockscreen(image string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	path := filepath.Join(usr.HomeDir, ".config", "kscreenlockerrc")
	content := fmt.Sprintf(`[Greeter][Wallpaper][org.kde.image][General]
Image=file://%s
PreviewImage=file://%s
`, image, image)

	return os.WriteFile(path, []byte(content), 0o644)
}
