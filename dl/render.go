package dl

import (
	"image/color"
	"math"

	"goforge.dev/scribe"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

// gstate is the PostScript-style graphics state.
type gstate struct {
	ctm geom.Affine
	col color.RGBA
	pen path.Pen
}

// Render interprets prog onto c. It is total: any Program renders
// without error (painting an empty path and Restore on an empty
// stack are no-ops).
func Render(prog Program, c *scribe.Canvas) {
	st := gstate{
		ctm: geom.Identity(),
		col: color.RGBA{A: 255},
		pen: path.Pen{Width: 1},
	}
	var stack []gstate
	var cur path.Path

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
		case SetColor:
			st.col = o.C
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
			c.Fill(&cur, st.col)
			cur = path.Path{}
		case FillEO:
			c.FillEvenOdd(&cur, st.col)
			cur = path.Path{}
		case Stroke:
			pen := st.pen
			// Stroke width is in user space: scale by the CTM's
			// average scale factor sqrt(|det|).
			pen.Width *= math.Sqrt(math.Abs(st.ctm.A*st.ctm.D - st.ctm.B*st.ctm.C))
			c.Stroke(&cur, pen, st.col)
			cur = path.Path{}
		case Clear:
			c.Clear(o.C)
		}
	}
}
