package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
)

func main() {
	dir := flag.String("d", "", "Directory to save image")
	noSave := flag.Bool("ns", false, "Do not save image permanently, use /tmp instead")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("Usage: wallpaper [-d dir] [-ns] <image-url>")
		return
	}
	imageURL := flag.Arg(0)

	var saveDir string
	if *noSave {
		saveDir = os.TempDir()
	} else if *dir != "" {
		saveDir = *dir
	} else {
		usr, _ := user.Current()
		saveDir = filepath.Join(usr.HomeDir, "wallpapers")
	}
	os.MkdirAll(saveDir, 0755)

	localPath, err := downloadAndNameImage(imageURL, saveDir)
	if err != nil {
		fmt.Println("Download failed:", err)
		return
	}
	if err := updateDesktop(localPath); err != nil {
		fmt.Println("Failed to set desktop wallpaper:", err)
	}
	if err := updateLockscreen(localPath); err != nil {
		fmt.Println("Failed to set lockscreen wallpaper:", err)
	}
}

func downloadAndNameImage(url, saveDir string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	ext := mimeToExt(contentType)
	name := extractFileName(url)
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

func extractFileName(url string) string {
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
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
	return os.WriteFile(path, []byte(content), 0644)
}
