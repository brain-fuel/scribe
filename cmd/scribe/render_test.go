package main

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderCommand(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "d.scr")
	scr := "#336699 setcolor\n2 2 12 12 3 continuous roundrect\nfill\n"
	if err := os.WriteFile(src, []byte(scr), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "d.png")
	if err := run([]string{"render", src, "-size", "16", "-o", out}); err != nil {
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
	if img.Bounds().Dx() != 16 || img.Bounds().Dy() != 16 {
		t.Errorf("bounds = %v", img.Bounds())
	}
	r, g, b, a := img.At(8, 8).RGBA()
	if a != 0xffff || r>>8 != 0x33 || g>>8 != 0x66 || b>>8 != 0x99 {
		t.Errorf("center = %x %x %x %x", r, g, b, a)
	}
}

func TestRenderSVG(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "d.scr")
	if err := os.WriteFile(src, []byte("#ff0000 setcolor\n0 0 8 8 rect\nfill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "d.svg")
	if err := run([]string{"render", src, "-size", "8", "-svg", out}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "<svg") || !strings.Contains(s, `fill="#ff0000"`) {
		t.Errorf("svg output wrong:\n%s", s)
	}
	// PNG default output must NOT appear when only -svg was asked.
	if _, err := os.Stat(filepath.Join(dir, "out.png")); err == nil {
		t.Error("default PNG written despite -svg-only request")
	}
}

func TestRenderParseErrorSurfaces(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bad.scr")
	if err := os.WriteFile(src, []byte("1 2 bogus\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"render", src, "-o", filepath.Join(dir, "x.png")})
	if err == nil {
		t.Fatal("want parse error")
	}
	if want := "1:5"; !strings.Contains(err.Error(), want) {
		t.Errorf("error %q lacks position %s", err, want)
	}
}
