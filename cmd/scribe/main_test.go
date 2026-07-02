package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestIconCommand(t *testing.T) {
	out := filepath.Join(t.TempDir(), "icon.png")
	err := run([]string{"icon", "-size", "64", "-fill", "#336699", "-o", out})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
		t.Errorf("bounds = %v", img.Bounds())
	}
	// corner pixel transparent, center opaque
	_, _, _, a := img.At(0, 0).RGBA()
	if a != 0 {
		t.Errorf("corner alpha = %d, want 0", a)
	}
	r, g, b, a := img.At(32, 32).RGBA()
	if a != 0xffff || r>>8 != 0x33 || g>>8 != 0x66 || b>>8 != 0x99 {
		t.Errorf("center = %x %x %x %x", r, g, b, a)
	}
}

func TestParseHexColor(t *testing.T) {
	c, err := parseHexColor("#FF6A00")
	if err != nil || c.R != 0xFF || c.G != 0x6A || c.B != 0x00 || c.A != 0xFF {
		t.Errorf("parse = %v, %v", c, err)
	}
	c, err = parseHexColor("#11223344")
	if err != nil || c.A != 0x44 {
		t.Errorf("parse rgba = %v, %v", c, err)
	}
	if _, err := parseHexColor("nope"); err == nil {
		t.Error("expected error")
	}
}
