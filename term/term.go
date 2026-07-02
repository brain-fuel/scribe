// Package term renders images to terminals: kitty graphics protocol,
// sixel, or half-block cells with 24-bit color. All encoders are
// deterministic: the same image produces the same bytes.
package term

import (
	"fmt"
	"image"
	"io"
	"strings"
)

// Protocol selects a terminal image encoding.
type Protocol uint8

const (
	// Auto picks via Detect at write time.
	Auto Protocol = iota
	// Kitty is the kitty graphics protocol (chunked base64 PNG).
	Kitty
	// Sixel is DEC sixel graphics.
	Sixel
	// HalfBlock renders two pixels per cell with the upper half
	// block glyph and 24-bit foreground/background colors.
	HalfBlock
)

// Detect guesses the best protocol from environment variables.
// getenv is injected for testability; pass os.Getenv.
func Detect(getenv func(string) string) Protocol {
	if getenv("KITTY_WINDOW_ID") != "" || getenv("TERM") == "xterm-kitty" {
		return Kitty
	}
	term := getenv("TERM")
	if term == "foot" || term == "mlterm" || strings.Contains(term, "sixel") ||
		getenv("WT_SESSION") != "" {
		return Sixel
	}
	return HalfBlock
}

// Write encodes img to w using protocol p.
func Write(w io.Writer, img image.Image, p Protocol) error {
	switch p {
	case Kitty:
		return writeKitty(w, img)
	case Sixel:
		return writeSixel(w, img)
	default:
		return writeHalfBlock(w, img)
	}
}

// overBlack composites a pixel over black, returning 8-bit channels.
func overBlack(img image.Image, x, y int) (r, g, b uint8) {
	r16, g16, b16, _ := img.At(x, y).RGBA()
	// RGBA() is alpha-premultiplied: over black is the value itself.
	return uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8)
}

func writeHalfBlock(w io.Writer, img image.Image) error {
	bounds := img.Bounds()
	var b []byte
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ur, ug, ub := overBlack(img, x, y)
			b = fmt.Appendf(b, "\x1b[38;2;%d;%d;%dm", ur, ug, ub)
			if y+1 < bounds.Max.Y {
				lr, lg, lb := overBlack(img, x, y+1)
				b = fmt.Appendf(b, "\x1b[48;2;%d;%d;%dm", lr, lg, lb)
			} else {
				b = append(b, "\x1b[49m"...)
			}
			b = append(b, "▀"...)
		}
		b = append(b, "\x1b[0m\n"...)
	}
	_, err := w.Write(b)
	return err
}

// Temporary stubs: replaced by the kitty and sixel encoder tasks.
func writeKitty(w io.Writer, img image.Image) error { return writeHalfBlock(w, img) }
func writeSixel(w io.Writer, img image.Image) error { return writeHalfBlock(w, img) }
