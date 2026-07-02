package path

import (
	"math"

	"goforge.dev/scribe/geom"
)

// Cap is a stroke end-cap style.
type Cap uint8

const (
	ButtCap Cap = iota
	RoundCap
	SquareCap
)

// Join is a stroke joint style.
type Join uint8

const (
	BevelJoin Join = iota
	RoundJoin
)

// Pen describes stroke geometry.
type Pen struct {
	Width float64
	Cap   Cap
	Join  Join
}

// Stroke converts the stroke of p under pen into a fill path. Render
// the result with the nonzero rule. Curves are flattened at tol.
//
// The construction is a nonzero union: one oriented quad per segment,
// a stamped kappa-table circle per round join or cap, a notch
// triangle per bevel join. Every contour is normalized to positive
// shoelace area so overlaps reinforce instead of cancel. No trig.
func Stroke(p *Path, pen Pen, tol float64) *Path {
	hw := pen.Width / 2
	var out Path
	if hw <= 0 {
		return &out
	}
	for _, pl := range p.Flatten(tol) {
		strokePolyline(&out, pl, hw, pen)
	}
	return &out
}

func strokePolyline(out *Path, pl Polyline, hw float64, pen Pen) {
	pts := dedupe(pl.Pts)
	if len(pts) == 0 {
		return
	}
	if len(pts) == 1 {
		switch pen.Cap {
		case RoundCap:
			stampCircle(out, pts[0], hw)
		case SquareCap:
			q := pts[0]
			addQuad(out,
				geom.Pt(q.X-hw, q.Y-hw), geom.Pt(q.X+hw, q.Y-hw),
				geom.Pt(q.X+hw, q.Y+hw), geom.Pt(q.X-hw, q.Y+hw))
		}
		return
	}
	n := len(pts)
	segEnd := n - 1
	if pl.Closed {
		segEnd = n // segment from last back to first
	}
	for i := 0; i < segEnd; i++ {
		a, b := pts[i], pts[(i+1)%n]
		t := b.Sub(a)
		l := math.Sqrt(t.X*t.X + t.Y*t.Y)
		u := geom.Pt(t.X/l, t.Y/l)
		nrm := geom.Pt(-u.Y, u.X)
		a2, b2 := a, b
		if !pl.Closed && pen.Cap == SquareCap {
			if i == 0 {
				a2 = a.Sub(u.Mul(hw))
			}
			if i == segEnd-1 {
				b2 = b.Add(u.Mul(hw))
			}
		}
		off := nrm.Mul(hw)
		addQuad(out, a2.Add(off), b2.Add(off), b2.Sub(off), a2.Sub(off))
	}
	// joins at interior vertices; for closed polylines every vertex
	joinStart, joinEnd := 1, n-1
	if pl.Closed {
		joinStart, joinEnd = 0, n
	}
	for i := joinStart; i < joinEnd; i++ {
		v := pts[i]
		switch pen.Join {
		case RoundJoin:
			stampCircle(out, v, hw)
		case BevelJoin:
			prev := pts[(i-1+n)%n]
			next := pts[(i+1)%n]
			bevel(out, prev, v, next, hw)
		}
	}
	// caps at open ends
	if !pl.Closed && pen.Cap == RoundCap {
		stampCircle(out, pts[0], hw)
		stampCircle(out, pts[n-1], hw)
	}
}

// dedupe removes consecutive duplicate points (and a duplicate wrap
// point on closed input).
func dedupe(pts []geom.Point) []geom.Point {
	out := pts[:0:0]
	for _, q := range pts {
		if len(out) == 0 || q != out[len(out)-1] {
			out = append(out, q)
		}
	}
	if len(out) > 1 && out[0] == out[len(out)-1] {
		out = out[:len(out)-1]
	}
	return out
}

// bevel fills the outer notch triangle at vertex v between the offset
// edges of segments prev-v and v-next.
func bevel(out *Path, prev, v, next geom.Point, hw float64) {
	t1 := v.Sub(prev)
	l1 := math.Sqrt(t1.X*t1.X + t1.Y*t1.Y)
	t2 := next.Sub(v)
	l2 := math.Sqrt(t2.X*t2.X + t2.Y*t2.Y)
	if l1 == 0 || l2 == 0 {
		return
	}
	n1 := geom.Pt(-t1.Y/l1, t1.X/l1)
	n2 := geom.Pt(-t2.Y/l2, t2.X/l2)
	// The notch is on the side away from the turn: cross > 0 means
	// the turn is toward +normal, so the gap is on the -normal side.
	cross := t1.X*t2.Y - t1.Y*t2.X
	if cross == 0 {
		return // collinear: no notch
	}
	s := hw
	if cross > 0 {
		s = -hw
	}
	addTri(out, v, v.Add(n1.Mul(s)), v.Add(n2.Mul(s)))
}

// stampCircle appends a positively-oriented circle contour.
func stampCircle(out *Path, c geom.Point, r float64) {
	k := kappa * r
	out.MoveTo(geom.Pt(c.X+r, c.Y))
	out.CubicTo(geom.Pt(c.X+r, c.Y+k), geom.Pt(c.X+k, c.Y+r), geom.Pt(c.X, c.Y+r))
	out.CubicTo(geom.Pt(c.X-k, c.Y+r), geom.Pt(c.X-r, c.Y+k), geom.Pt(c.X-r, c.Y))
	out.CubicTo(geom.Pt(c.X-r, c.Y-k), geom.Pt(c.X-k, c.Y-r), geom.Pt(c.X, c.Y-r))
	out.CubicTo(geom.Pt(c.X+k, c.Y-r), geom.Pt(c.X+r, c.Y-k), geom.Pt(c.X+r, c.Y))
	out.Close()
}

// addQuad appends the quad contour normalized to positive shoelace.
func addQuad(out *Path, a, b, c, d geom.Point) {
	if shoelace4(a, b, c, d) < 0 {
		a, b, c, d = d, c, b, a
	}
	out.MoveTo(a)
	out.LineTo(b)
	out.LineTo(c)
	out.LineTo(d)
	out.Close()
}

// addTri appends the triangle contour normalized to positive shoelace.
func addTri(out *Path, a, b, c geom.Point) {
	if (b.X-a.X)*(c.Y-a.Y)-(c.X-a.X)*(b.Y-a.Y) < 0 {
		b, c = c, b
	}
	out.MoveTo(a)
	out.LineTo(b)
	out.LineTo(c)
	out.Close()
}

func shoelace4(a, b, c, d geom.Point) float64 {
	return (a.X*b.Y - b.X*a.Y) + (b.X*c.Y - c.X*b.Y) +
		(c.X*d.Y - d.X*c.Y) + (d.X*a.Y - a.X*d.Y)
}
