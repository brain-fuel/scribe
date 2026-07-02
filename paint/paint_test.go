package paint

import (
	"image/color"
	"testing"
)

func TestSolid(t *testing.T) {
	s := Solid(color.RGBA{R: 255, A: 255})
	r, g, b, a := s.At(100, -5).RGBA()
	if r != 0xffff || g != 0 || b != 0 || a != 0xffff {
		t.Errorf("Solid.At = %v %v %v %v", r, g, b, a)
	}
}
