package term

import (
	"bytes"
	"image"
	"image/color"
	"strings"
	"testing"
)

func img2x2() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, 2, 2))
	m.SetRGBA(0, 0, color.RGBA{255, 0, 0, 255})
	m.SetRGBA(1, 0, color.RGBA{0, 255, 0, 255})
	m.SetRGBA(0, 1, color.RGBA{0, 0, 255, 255})
	m.SetRGBA(1, 1, color.RGBA{255, 255, 255, 255})
	return m
}

func TestHalfBlock2x2(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, img2x2(), HalfBlock); err != nil {
		t.Fatal(err)
	}
	want := "\x1b[38;2;255;0;0m\x1b[48;2;0;0;255m▀" +
		"\x1b[38;2;0;255;0m\x1b[48;2;255;255;255m▀" +
		"\x1b[0m\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestHalfBlockOddHeight(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 1, 1))
	m.SetRGBA(0, 0, color.RGBA{10, 20, 30, 255})
	var buf bytes.Buffer
	if err := Write(&buf, m, HalfBlock); err != nil {
		t.Fatal(err)
	}
	want := "\x1b[38;2;10;20;30m\x1b[49m▀\x1b[0m\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

// Transparency composites over black.
func TestHalfBlockAlpha(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 1, 2))
	// Premultiplied half-transparent white: (128,128,128,128).
	m.SetRGBA(0, 0, color.RGBA{128, 128, 128, 128})
	m.SetRGBA(0, 1, color.RGBA{0, 0, 0, 0})
	var buf bytes.Buffer
	if err := Write(&buf, m, HalfBlock); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "38;2;128;128;128m") {
		t.Errorf("premultiplied color should pass through over black: %q", out)
	}
	if !strings.Contains(out, "48;2;0;0;0m") {
		t.Errorf("transparent should be black: %q", out)
	}
}

func TestDetect(t *testing.T) {
	cases := []struct {
		env  map[string]string
		want Protocol
	}{
		{map[string]string{"KITTY_WINDOW_ID": "1"}, Kitty},
		{map[string]string{"TERM": "xterm-kitty"}, Kitty},
		{map[string]string{"TERM": "foot", "COLORTERM": "truecolor"}, Sixel},
		{map[string]string{"WT_SESSION": "abc"}, Sixel},
		{map[string]string{"TERM": "xterm-256color"}, HalfBlock},
		{map[string]string{}, HalfBlock},
	}
	for _, tc := range cases {
		got := Detect(func(k string) string { return tc.env[k] })
		if got != tc.want {
			t.Errorf("Detect(%v) = %v, want %v", tc.env, got, tc.want)
		}
	}
}
