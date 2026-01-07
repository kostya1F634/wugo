package image

import (
	"bytes"
	"testing"
)

func TestRandomHexLength(t *testing.T) {
	got := randomHex(5, bytes.NewReader([]byte{0x0a, 0x0b, 0x0c}))
	if got != "0a0b0" {
		t.Fatalf("unexpected hex: %s", got)
	}
}
