package geom

import (
	"image"
	"testing"
)

func TestPointOps(t *testing.T) {
	p := Pt(1, 2)
	q := Pt(3, 5)
	if got := p.Add(q); got != Pt(4, 7) {
		t.Errorf("Add = %v", got)
	}
	if got := q.Sub(p); got != Pt(2, 3) {
		t.Errorf("Sub = %v", got)
	}
	if got := p.Mul(2); got != Pt(2, 4) {
		t.Errorf("Mul = %v", got)
	}
	if got := p.Lerp(q, 0.5); got != Pt(2, 3.5) {
		t.Errorf("Lerp = %v", got)
	}
}

// Corner-grid semantics: Rect(0,0,4,4) covers exactly the 4x4 pixel block.
// Width is exactly Max.X - Min.X. No off-by-one.
func TestRectCornerGrid(t *testing.T) {
	r := RectXYWH(0, 0, 4, 4)
	if r.W() != 4 || r.H() != 4 {
		t.Fatalf("W,H = %v,%v", r.W(), r.H())
	}
	if got, want := r.OuterPixels(), image.Rect(0, 0, 4, 4); got != want {
		t.Errorf("OuterPixels = %v, want %v", got, want)
	}
	// A rect whose edges fall mid-pixel touches the pixels it overlaps.
	r2 := RectXYWH(0.5, 0.5, 1, 1) // covers points (0.5,0.5)-(1.5,1.5)
	if got, want := r2.OuterPixels(), image.Rect(0, 0, 2, 2); got != want {
		t.Errorf("OuterPixels = %v, want %v", got, want)
	}
}

func TestAffine(t *testing.T) {
	m := Translate(10, 20)
	if got := m.Apply(Pt(1, 1)); got != Pt(11, 21) {
		t.Errorf("Translate.Apply = %v", got)
	}
	s := Scale(2, 3)
	if got := s.Apply(Pt(1, 1)); got != Pt(2, 3) {
		t.Errorf("Scale.Apply = %v", got)
	}
	// m.Mul(n) applies n first, then m (standard matrix composition).
	ts := Translate(10, 0).Mul(Scale(2, 2))
	if got := ts.Apply(Pt(1, 1)); got != Pt(12, 2) {
		t.Errorf("compose = %v, want (12,2)", got)
	}
	if got := Identity().Apply(Pt(7, 8)); got != Pt(7, 8) {
		t.Errorf("Identity.Apply = %v", got)
	}
}
