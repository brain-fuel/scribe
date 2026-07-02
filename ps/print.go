package ps

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

// Print emits the canonical ps text for prog: one op per line.
// Parse(Print(p)) reproduces p exactly.
func Print(prog dl.Program) string {
	var b strings.Builder
	for _, op := range prog {
		b.WriteString(opLine(op))
		b.WriteByte('\n')
	}
	return b.String()
}

func num(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

func pt(p geom.Point) string { return num(p.X) + " " + num(p.Y) }

func col(c color.RGBA) string {
	if c.A == 0xFF {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}

// rectArgs prints x y w h. Note w and h are Max minus Min; Parse
// rebuilds via RectXYWH (Min plus size). For coordinates exactly
// representable in float64 this round-trips exactly.
func rectArgs(r geom.Rect) string {
	return pt(r.Min) + " " + num(r.W()) + " " + num(r.H())
}

func opLine(op dl.Op) string {
	switch o := op.(type) {
	case dl.MoveTo:
		return pt(o.P) + " moveto"
	case dl.LineTo:
		return pt(o.P) + " lineto"
	case dl.CubicTo:
		return pt(o.C1) + " " + pt(o.C2) + " " + pt(o.P) + " curveto"
	case dl.ClosePath:
		return "closepath"
	case dl.NewPath:
		return "newpath"
	case dl.Rect:
		return rectArgs(o.R) + " rect"
	case dl.Circle:
		return pt(o.C) + " " + num(o.Radius) + " circle"
	case dl.RoundRect:
		style := "circular"
		if o.Style == path.Continuous {
			style = "continuous"
		}
		return rectArgs(o.R) + " " + num(o.Radius) + " " + style + " roundrect"
	case dl.SetColor:
		return col(o.C) + " setcolor"
	case dl.SetLineWidth:
		return num(o.W) + " setlinewidth"
	case dl.SetCap:
		switch o.C {
		case path.RoundCap:
			return "round setcap"
		case path.SquareCap:
			return "square setcap"
		default:
			return "butt setcap"
		}
	case dl.SetJoin:
		if o.J == path.RoundJoin {
			return "round setjoin"
		}
		return "bevel setjoin"
	case dl.Save:
		return "gsave"
	case dl.Restore:
		return "grestore"
	case dl.Translate:
		return num(o.X) + " " + num(o.Y) + " translate"
	case dl.Scale:
		return num(o.X) + " " + num(o.Y) + " scale"
	case dl.Fill:
		return "fill"
	case dl.FillEO:
		return "eofill"
	case dl.Stroke:
		return "stroke"
	case dl.Clear:
		return col(o.C) + " clear"
	}
	panic(fmt.Sprintf("ps: unknown op %T", op))
}
