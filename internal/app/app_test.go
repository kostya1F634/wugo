package app

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

type fakeProcessor struct {
	input   string
	saveDir string
	noMove  bool
	result  string
	err     error
}

func (f *fakeProcessor) Process(_ context.Context, input, saveDir string, noMove bool) (string, error) {
	f.input = input
	f.saveDir = saveDir
	f.noMove = noMove
	return f.result, f.err
}

type fakeSetter struct {
	desktopCalls int
	lockCalls    int
	desktopErr   error
	lockErr      error
}

func (f *fakeSetter) SetDesktop(_ context.Context, _ string) error {
	f.desktopCalls++
	return f.desktopErr
}

func (f *fakeSetter) SetLockscreen(_ context.Context, _ string) error {
	f.lockCalls++
	return f.lockErr
}

func TestMainUsage(t *testing.T) {
	var out bytes.Buffer
	deps := Deps{
		Processor: &fakeProcessor{},
		Setter:    &fakeSetter{},
		Out:       &out,
		Err:       &out,
		MkdirAll:  func(string, fs.FileMode) error { return nil },
		HomeDir:   func() (string, error) { return "/tmp", nil },
	}

	code := Main(context.Background(), nil, deps)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(out.String(), "Usage: wugo") {
		t.Fatalf("expected usage output, got %s", out.String())
	}
}

func TestMainSuccess(t *testing.T) {
	var out bytes.Buffer
	processor := &fakeProcessor{result: "/tmp/image.png"}
	setter := &fakeSetter{}
	var mkdirPath string
	deps := Deps{
		Processor: processor,
		Setter:    setter,
		Out:       &out,
		Err:       &out,
		MkdirAll: func(path string, _ fs.FileMode) error {
			mkdirPath = path
			return nil
		},
		HomeDir: func() (string, error) { return "/home/test", nil },
	}

	code := Main(context.Background(), []string{"/tmp/input.png"}, deps)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if setter.desktopCalls != 1 || setter.lockCalls != 1 {
		t.Fatalf("expected setters called, got %d/%d", setter.desktopCalls, setter.lockCalls)
	}
	expectedDir := filepath.Join("/home/test", "wallpapers")
	if mkdirPath != expectedDir {
		t.Fatalf("expected mkdir %s, got %s", expectedDir, mkdirPath)
	}
	if !strings.Contains(out.String(), "Wallpaper set successfully") {
		t.Fatalf("expected success output, got %s", out.String())
	}
}

func TestMainProcessorError(t *testing.T) {
	var out bytes.Buffer
	processor := &fakeProcessor{err: errors.New("boom")}
	deps := Deps{
		Processor: processor,
		Setter:    &fakeSetter{},
		Out:       &out,
		Err:       &out,
		MkdirAll:  func(string, fs.FileMode) error { return nil },
		HomeDir:   func() (string, error) { return "/home/test", nil },
	}

	code := Main(context.Background(), []string{"/tmp/input.png"}, deps)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(out.String(), "Failed to process image") {
		t.Fatalf("expected process error output, got %s", out.String())
	}
}

func TestMainSetterError(t *testing.T) {
	var out bytes.Buffer
	processor := &fakeProcessor{result: "/tmp/image.png"}
	setter := &fakeSetter{desktopErr: errors.New("dbus")}
	deps := Deps{
		Processor: processor,
		Setter:    setter,
		Out:       &out,
		Err:       &out,
		MkdirAll:  func(string, fs.FileMode) error { return nil },
		HomeDir:   func() (string, error) { return "/home/test", nil },
	}

	code := Main(context.Background(), []string{"/tmp/input.png"}, deps)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if setter.desktopCalls != 1 || setter.lockCalls != 1 {
		t.Fatalf("expected setters called, got %d/%d", setter.desktopCalls, setter.lockCalls)
	}
	if !strings.Contains(out.String(), "Failed to set desktop wallpaper") {
		t.Fatalf("expected desktop error output, got %s", out.String())
	}
}
