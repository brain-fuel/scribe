package dl

import (
	"testing"

	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

func TestMasksPerPaintOp(t *testing.T) {
	prog := Program{
		Clear{},
		RoundRect{geom.RectXYWH(2, 2, 28, 28), 6, path.Continuous},
		Fill{},
		Circle{geom.Pt(16, 16), 8},
		FillEO{},
	}
	ms := Masks(prog, 32, 32)
	if len(ms) != 2 {
		t.Fatalf("masks = %d, want 2 (one per paint op)", len(ms))
	}
	if got := ms[0].AlphaAt(16, 16).A; got != 255 {
		t.Errorf("roundrect center = %d, want 255", got)
	}
	if got := ms[1].AlphaAt(16, 16).A; got != 255 {
		t.Errorf("circle center = %d, want 255", got)
	}
	if got := ms[1].AlphaAt(2, 2).A; got != 0 {
		t.Errorf("circle corner = %d, want 0", got)
	}
}

// A transform in force at paint time shapes the mask.
func TestMasksTransform(t *testing.T) {
	prog := Program{
		Save{},
		Translate{4, 4},
		Scale{2, 2},
		Rect{geom.RectXYWH(0, 0, 2, 2)},
		Fill{},
		Restore{},
	}
	ms := Masks(prog, 16, 16)
	if len(ms) != 1 {
		t.Fatal("want 1 mask")
	}
	if got := ms[0].AlphaAt(5, 5).A; got != 255 {
		t.Errorf("(5,5) = %d, want 255 (device rect 4..8)", got)
	}
	if got := ms[0].AlphaAt(9, 9).A; got != 0 {
		t.Errorf("(9,9) = %d, want 0", got)
	}
}
