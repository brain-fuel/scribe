package dl

import (
	"image"
	"math"

	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
	"goforge.dev/scribe/raster"

	"goforge.dev/scribe"
)

// Masks interprets prog like Render but, instead of compositing, returns the
// raw coverage mask of each paint op (Fill, FillEO, Stroke) in program
// order. Clear and color/state ops produce no mask. This is the semantic
// surface the cross-language divergence lock compares: PNG and compositing
// are presentation, masks are meaning.
func Masks(prog Program, w, h int) []*image.Alpha {
	st := gstate{ctm: geom.Identity(), pen: path.Pen{Width: 1}}
	var stack []gstate
	var cur path.Path
	var out []*image.Alpha

	appendShape := func(sp *path.Path) {
		pts := sp.Points()
		pi := 0
		for _, v := range sp.Verbs() {
			switch v {
			case path.VMoveTo:
				cur.MoveTo(st.ctm.Apply(pts[pi]))
				pi++
			case path.VLineTo:
				cur.LineTo(st.ctm.Apply(pts[pi]))
				pi++
			case path.VCubicTo:
				cur.CubicTo(st.ctm.Apply(pts[pi]), st.ctm.Apply(pts[pi+1]), st.ctm.Apply(pts[pi+2]))
				pi += 3
			case path.VClose:
				cur.Close()
			}
		}
	}
	rasterizeCur := func(p *path.Path, rule raster.FillRule) {
		r := raster.New(w, h)
		for _, pl := range p.Flatten(scribe.FlattenTol) {
			r.AddPolyline(pl)
		}
		out = append(out, r.Mask(rule))
	}

	for _, op := range prog {
		switch o := op.(type) {
		case MoveTo:
			cur.MoveTo(st.ctm.Apply(o.P))
		case LineTo:
			cur.LineTo(st.ctm.Apply(o.P))
		case CubicTo:
			cur.CubicTo(st.ctm.Apply(o.C1), st.ctm.Apply(o.C2), st.ctm.Apply(o.P))
		case ClosePath:
			cur.Close()
		case NewPath:
			cur = path.Path{}
		case Rect:
			appendShape(path.RectPath(o.R))
		case Circle:
			appendShape(path.Circle(o.C, o.Radius))
		case RoundRect:
			appendShape(path.RoundRect(o.R, o.Radius, o.Style))
		case SetLineWidth:
			st.pen.Width = o.W
		case SetCap:
			st.pen.Cap = o.C
		case SetJoin:
			st.pen.Join = o.J
		case Save:
			stack = append(stack, st)
		case Restore:
			if n := len(stack); n > 0 {
				st = stack[n-1]
				stack = stack[:n-1]
			}
		case Translate:
			st.ctm = st.ctm.Mul(geom.Translate(o.X, o.Y))
		case Scale:
			st.ctm = st.ctm.Mul(geom.Scale(o.X, o.Y))
		case Fill:
			rasterizeCur(&cur, raster.NonZero)
			cur = path.Path{}
		case FillEO:
			rasterizeCur(&cur, raster.EvenOdd)
			cur = path.Path{}
		case Stroke:
			pen := st.pen
			pen.Width *= math.Sqrt(math.Abs(st.ctm.A*st.ctm.D - st.ctm.B*st.ctm.C))
			rasterizeCur(path.Stroke(&cur, pen, scribe.FlattenTol), raster.NonZero)
			cur = path.Path{}
		}
	}
	return out
}
