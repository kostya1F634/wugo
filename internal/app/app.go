package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"wugo/internal/wallpaper"
)

var ErrUsage = errors.New("usage")

type Options struct {
	SaveDir string
	NoMove  bool
}

type ImageProcessor interface {
	Process(ctx context.Context, input, saveDir string, noMove bool) (string, error)
}

type Deps struct {
	Processor ImageProcessor
	Setter    wallpaper.Setter
	Out       io.Writer
	Err       io.Writer
	MkdirAll  func(path string, perm fs.FileMode) error
	HomeDir   func() (string, error)
}

func Main(ctx context.Context, args []string, deps Deps) int {
	deps = withDefaults(deps)

	opts, input, err := ParseArgs(args)
	if err != nil {
		if errors.Is(err, ErrUsage) {
			usage(deps.Err)
			return 2
		}
		fmt.Fprintln(deps.Err, err)
		usage(deps.Err)
		return 2
	}

	saveDir, err := resolveSaveDir(opts.SaveDir, deps.HomeDir)
	if err != nil {
		fmt.Fprintln(deps.Err, "Failed to resolve save directory:", err)
		return 1
	}

	if err := deps.MkdirAll(saveDir, 0o755); err != nil {
		fmt.Fprintln(deps.Err, "Failed to create directory:", err)
		return 1
	}

	localPath, err := deps.Processor.Process(ctx, input, saveDir, opts.NoMove)
	if err != nil {
		fmt.Fprintln(deps.Err, "Failed to process image:", err)
		return 1
	}

	hadErr := false
	if err := deps.Setter.SetDesktop(ctx, localPath); err != nil {
		fmt.Fprintln(deps.Err, "Failed to set desktop wallpaper:", err)
		hadErr = true
	}
	if err := deps.Setter.SetLockscreen(ctx, localPath); err != nil {
		fmt.Fprintln(deps.Err, "Failed to set lockscreen wallpaper:", err)
		hadErr = true
	}

	if hadErr {
		return 1
	}

	fmt.Fprintln(deps.Out, "Wallpaper set successfully:", localPath)
	return 0
}

func ParseArgs(args []string) (Options, string, error) {
	fs := flag.NewFlagSet("wugo", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dir := fs.String("d", "", "Directory to save/move image")
	noMove := fs.Bool("nm", false, "Do not move local file, use it from current location")

	if err := fs.Parse(args); err != nil {
		return Options{}, "", fmt.Errorf("parse flags: %w", err)
	}

	if fs.NArg() < 1 {
		return Options{}, "", ErrUsage
	}

	return Options{SaveDir: *dir, NoMove: *noMove}, fs.Arg(0), nil
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "Usage: wugo [-d dir] [-nm] <image-url-or-path>")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  -d   Directory to save/move image (default: ~/wallpapers)")
	fmt.Fprintln(w, "  -nm  Do not move local file, use it from current location")
}

func resolveSaveDir(dir string, homeDir func() (string, error)) (string, error) {
	if dir == "" {
		home, err := homeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, "wallpapers")
	}

	return filepath.Abs(dir)
}

func withDefaults(deps Deps) Deps {
	if deps.Out == nil {
		deps.Out = io.Discard
	}
	if deps.Err == nil {
		deps.Err = io.Discard
	}
	if deps.MkdirAll == nil {
		deps.MkdirAll = os.MkdirAll
	}
	if deps.HomeDir == nil {
		deps.HomeDir = os.UserHomeDir
	}

	return deps
}
