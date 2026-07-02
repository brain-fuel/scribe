package term

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestKittyRoundTrip(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 3, 2))
	m.SetRGBA(1, 1, color.RGBA{9, 8, 7, 255})
	var buf bytes.Buffer
	if err := Write(&buf, m, Kitty); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "\x1b_G") || !strings.HasSuffix(out, "\x1b\\") {
		t.Fatalf("not an APC escape: %q", out)
	}
	// Single chunk expected for a tiny image: keys then payload.
	body := strings.TrimSuffix(strings.TrimPrefix(out, "\x1b_G"), "\x1b\\")
	semi := strings.IndexByte(body, ';')
	if semi < 0 {
		t.Fatalf("no key/payload separator: %q", out)
	}
	keys, payload := body[:semi], body[semi+1:]
	for _, k := range []string{"a=T", "f=100", "m=0"} {
		if !strings.Contains(keys, k) {
			t.Errorf("keys %q missing %s", keys, k)
		}
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatal(err)
	}
	back, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if back.Bounds().Dx() != 3 || back.Bounds().Dy() != 2 {
		t.Errorf("decoded bounds = %v", back.Bounds())
	}
}

// Payloads over 4096 base64 bytes must chunk with m=1 continuations.
func TestKittyChunking(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 128, 128))
	// Deterministic LCG noise: repeating patterns compress under the
	// 4096-byte chunk size; noise does not.
	s := uint32(1)
	for i := range m.Pix {
		s = s*1664525 + 1013904223
		m.Pix[i] = uint8(s >> 24)
	}
	var buf bytes.Buffer
	if err := Write(&buf, m, Kitty); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	chunks := strings.Count(out, "\x1b_G")
	if chunks < 2 {
		t.Fatalf("expected multiple chunks, got %d", chunks)
	}
	if strings.Count(out, "m=1") != chunks-1 {
		t.Errorf("want %d continuation chunks, got %d", chunks-1, strings.Count(out, "m=1"))
	}
	if !strings.Contains(out, "m=0") {
		t.Error("missing final m=0 chunk")
	}
}
