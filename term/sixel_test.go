package term

import (
	"bytes"
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestSixelStructure(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 4, 7)) // spans two bands
	for y := 0; y < 7; y++ {
		for x := 0; x < 4; x++ {
			m.SetRGBA(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := Write(&buf, m, Sixel); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "\x1bPq\"1;1;4;7") {
		t.Fatalf("bad intro: %q", out)
	}
	if !strings.HasSuffix(out, "\x1b\\") {
		t.Fatalf("bad terminator: %q", out)
	}
	// Pure red: level (255*5+127)/255 = 5 -> index 5*36 = 180,
	// defined as 100;0;0 percent.
	if !strings.Contains(out, "#180;2;100;0;0") {
		t.Errorf("missing palette def: %q", out)
	}
	// Band 1: all 6 rows set -> bitmask 63 -> char 126 '~', run of 4
	// columns -> RLE "!4~".
	if !strings.Contains(out, "!4~") {
		t.Errorf("missing RLE full band: %q", out)
	}
	if !strings.Contains(out, "-") {
		t.Errorf("missing band separator: %q", out)
	}
	// Band 2: only row 0 set -> char 63+1 = '@'.
	if !strings.Contains(out, "!4@") {
		t.Errorf("missing second band data: %q", out)
	}
}

func TestSixelDeterminism(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 9, 9))
	for i := range m.Pix {
		m.Pix[i] = uint8(i * 13)
	}
	render := func() string {
		var buf bytes.Buffer
		if err := Write(&buf, m, Sixel); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}
	if render() != render() {
		t.Error("sixel output not deterministic")
	}
}

// Two colors in one band both appear, in ascending register order,
// with a carriage return between passes.
func TestSixelTwoColors(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 2, 1))
	m.SetRGBA(0, 0, color.RGBA{255, 0, 0, 255}) // index 180
	m.SetRGBA(1, 0, color.RGBA{0, 0, 255, 255}) // index 5
	var buf bytes.Buffer
	if err := Write(&buf, m, Sixel); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	blue := strings.Index(out, "#5;2;0;0;100")
	red := strings.Index(out, "#180;2;100;0;0")
	if blue < 0 || red < 0 || blue > red {
		t.Errorf("palette defs missing or out of order: %q", out)
	}
	if !strings.Contains(out, "$") {
		t.Errorf("missing carriage return between color passes: %q", out)
	}
}
