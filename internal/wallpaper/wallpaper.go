package wallpaper

import (
	"context"
	"io/fs"
)

type Setter interface {
	SetDesktop(ctx context.Context, imagePath string) error
	SetLockscreen(ctx context.Context, imagePath string) error
}

type ScriptRunner interface {
	Run(script string) error
}

type FileWriter interface {
	WriteFile(name string, data []byte, perm fs.FileMode) error
}
