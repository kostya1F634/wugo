package wallpaper

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type KDESetter struct {
	Runner  ScriptRunner
	Writer  FileWriter
	HomeDir func() (string, error)
}

func NewKDESetter() *KDESetter {
	return &KDESetter{
		Runner:  DBusRunner{},
		Writer:  OSFileWriter{},
		HomeDir: os.UserHomeDir,
	}
}

func (k *KDESetter) SetDesktop(_ context.Context, imagePath string) error {
	k.applyDefaults()

	uri := fileURI(imagePath)
	quotedURI := strconv.Quote(uri)
	// Script API: https://develop.kde.org/docs/plasma/scripting/
	script := fmt.Sprintf(`var d = desktops(); for (let i in d) {
	d[i].wallpaperPlugin = "org.kde.image";
	d[i].currentConfigGroup = ["Wallpaper", "org.kde.image", "General"];
	d[i].writeConfig("Image", %s);
}`, quotedURI)

	return k.Runner.Run(script)
}

func (k *KDESetter) SetLockscreen(_ context.Context, imagePath string) error {
	k.applyDefaults()

	home, err := k.HomeDir()
	if err != nil {
		return err
	}

	uri := fileURI(imagePath)
	path := filepath.Join(home, ".config", "kscreenlockerrc")
	content := fmt.Sprintf(`[Greeter][Wallpaper][org.kde.image][General]
Image=%s
PreviewImage=%s
`, uri, uri)

	return k.Writer.WriteFile(path, []byte(content), 0o644)
}

func (k *KDESetter) applyDefaults() {
	if k.Runner == nil {
		k.Runner = DBusRunner{}
	}
	if k.Writer == nil {
		k.Writer = OSFileWriter{}
	}
	if k.HomeDir == nil {
		k.HomeDir = os.UserHomeDir
	}
}

func fileURI(path string) string {
	uri := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	return uri.String()
}
