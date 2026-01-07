package wallpaper

import (
	"io/fs"
	"os"

	"github.com/godbus/dbus/v5"
)

type DBusRunner struct{}

func (DBusRunner) Run(script string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	obj := conn.Object("org.kde.plasmashell", "/PlasmaShell")
	return obj.Call("org.kde.PlasmaShell.evaluateScript", 0, script).Err
}

type OSFileWriter struct{}

func (OSFileWriter) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
