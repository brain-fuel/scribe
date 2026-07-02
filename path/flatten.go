package path

import "goforge.dev/scribe/geom"

// Polyline is a flattened subpath. Closed polylines do not repeat the
// first point.
type Polyline struct {
	Pts    []geom.Point
	Closed bool
}

// Flatten converts p to polylines whose maximum deviation from the
// true curves is tol.
func (p *Path) Flatten(tol float64) []Polyline {
	var out []Polyline
	var cur []geom.Point
	flush := func(closed bool) {
		if len(cur) > 0 {
			out = append(out, Polyline{Pts: cur, Closed: closed})
		}
		cur = nil
	}
	pi := 0
	for _, v := range p.verbs {
		switch v {
		case VMoveTo:
			flush(false)
			cur = append(cur, p.pts[pi])
			pi++
		case VLineTo:
			cur = append(cur, p.pts[pi])
			pi++
		case VCubicTo:
			p0 := cur[len(cur)-1]
			cur = flattenCubic(cur, p0, p.pts[pi], p.pts[pi+1], p.pts[pi+2], tol, 0)
			pi += 3
		case VClose:
			flush(true)
		}
	}
	flush(false)
	return out
}

const maxDepth = 24

// flattenCubic appends line-approximation points for the cubic
// (p0, c1, c2, p3) to out, excluding p0, including p3.
func flattenCubic(out []geom.Point, p0, c1, c2, p3 geom.Point, tol float64, depth int) []geom.Point {
	if depth >= maxDepth || cubicFlat(p0, c1, c2, p3, tol) {
		return append(out, p3)
	}
	// de Casteljau split at t = 0.5
	ab := p0.Lerp(c1, 0.5)
	bc := c1.Lerp(c2, 0.5)
	cd := c2.Lerp(p3, 0.5)
	abc := ab.Lerp(bc, 0.5)
	bcd := bc.Lerp(cd, 0.5)
	mid := abc.Lerp(bcd, 0.5)
	out = flattenCubic(out, p0, ab, abc, mid, tol, depth+1)
	return flattenCubic(out, mid, bcd, cd, p3, tol, depth+1)
}

// cubicFlat reports whether both control points lie within tol of the
// chord p0-p3.
func cubicFlat(p0, c1, c2, p3 geom.Point, tol float64) bool {
	d := p3.Sub(p0)
	len2 := d.X*d.X + d.Y*d.Y
	if len2 < 1e-30 {
		// Degenerate chord: measure control point distance from p0.
		return dist2(c1, p0) <= tol*tol && dist2(c2, p0) <= tol*tol
	}
	// Perpendicular distance of c to line p0-p3 is |cross(d, c-p0)| / |d|.
	t2 := tol * tol * len2
	c1v := c1.Sub(p0)
	c2v := c2.Sub(p0)
	x1 := d.X*c1v.Y - d.Y*c1v.X
	x2 := d.X*c2v.Y - d.Y*c2v.X
	return x1*x1 <= t2 && x2*x2 <= t2
}

func dist2(a, b geom.Point) float64 {
	dx, dy := a.X-b.X, a.Y-b.Y
	return dx*dx + dy*dy
}
