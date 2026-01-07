package wallpaper

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	script string
	err    error
}

func (f *fakeRunner) Run(script string) error {
	f.script = script
	return f.err
}

type fakeWriter struct {
	path string
	data []byte
	perm fs.FileMode
	err  error
}

func (f *fakeWriter) WriteFile(name string, data []byte, perm fs.FileMode) error {
	f.path = name
	f.data = data
	f.perm = perm
	return f.err
}

func TestFileURI(t *testing.T) {
	got := fileURI("/home/user/My Picture.png")
	if got != "file:///home/user/My%20Picture.png" {
		t.Fatalf("unexpected URI: %s", got)
	}
}

func TestSetDesktopScript(t *testing.T) {
	runner := &fakeRunner{}
	setter := &KDESetter{Runner: runner}

	if err := setter.SetDesktop(context.Background(), "/home/user/My Picture.png"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(runner.script, "writeConfig(\"Image\", \"file:///home/user/My%20Picture.png\")") {
		t.Fatalf("script missing file URI: %s", runner.script)
	}
}

func TestSetLockscreenWritesConfig(t *testing.T) {
	runner := &fakeRunner{}
	writer := &fakeWriter{}
	home := t.TempDir()
	setter := &KDESetter{
		Runner:  runner,
		Writer:  writer,
		HomeDir: func() (string, error) { return home, nil },
	}

	if err := setter.SetLockscreen(context.Background(), "/home/user/My Picture.png"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(home, ".config", "kscreenlockerrc")
	if writer.path != expectedPath {
		t.Fatalf("unexpected path: %s", writer.path)
	}
	if writer.perm != 0o644 {
		t.Fatalf("unexpected perm: %v", writer.perm)
	}

	content := string(writer.data)
	if !strings.Contains(content, "Image=file:///home/user/My%20Picture.png") {
		t.Fatalf("missing image entry: %s", content)
	}
	if !strings.Contains(content, "PreviewImage=file:///home/user/My%20Picture.png") {
		t.Fatalf("missing preview entry: %s", content)
	}
}
