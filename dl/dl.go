// Package dl is scribe's display list: drawing operations as plain
// data. A Program is the portable intermediate representation of a
// drawing. It is deterministic and serializable (package ps is the
// canonical text codec), and each backend interprets it
// independently. Shape ops stay semantic (a RoundRect stays a
// RoundRect with its corner style) so backends and ports see meaning.
package dl

import (
	"image/color"

	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

// Program is a sequence of drawing operations.
type Program []Op

// Op is a display-list operation. The set is sealed.
type Op interface{ isOp() }

// Path construction. Points are in user space; the CTM applies when
// the op executes.
type (
	// MoveTo starts a new subpath at P.
	MoveTo struct{ P geom.Point }
	// LineTo appends a line to P.
	LineTo struct{ P geom.Point }
	// CubicTo appends a cubic Bezier via C1, C2 to P.
	CubicTo struct{ C1, C2, P geom.Point }
	// ClosePath closes the current subpath.
	ClosePath struct{}
	// NewPath discards the current path.
	NewPath struct{}
)

// Shapes: appended to the current path as closed subpaths.
type (
	// Rect appends a rectangle.
	Rect struct{ R geom.Rect }
	// Circle appends a circle (kappa table).
	Circle struct {
		C      geom.Point
		Radius float64
	}
	// RoundRect appends a rounded rectangle with the given corner
	// style (corner tables).
	RoundRect struct {
		R      geom.Rect
		Radius float64
		Style  path.CornerStyle
	}
)

// Graphics state.
type (
	// SetColor sets the paint color.
	SetColor struct{ C color.RGBA }
	// SetLineWidth sets the stroke width in user space.
	SetLineWidth struct{ W float64 }
	// SetCap sets the stroke cap.
	SetCap struct{ C path.Cap }
	// SetJoin sets the stroke join.
	SetJoin struct{ J path.Join }
	// Save pushes the graphics state (CTM, color, pen).
	Save struct{}
	// Restore pops the graphics state. No-op if the stack is empty.
	Restore struct{}
	// Translate concatenates a translation onto the CTM.
	Translate struct{ X, Y float64 }
	// Scale concatenates a scale onto the CTM.
	Scale struct{ X, Y float64 }
)

// Painting. Fill, FillEO and Stroke consume the current path.
type (
	// Fill paints the current path with the nonzero rule.
	Fill struct{}
	// FillEO paints the current path with the even-odd rule.
	FillEO struct{}
	// Stroke strokes the current path.
	Stroke struct{}
	// Clear fills the whole canvas with a color.
	Clear struct{ C color.RGBA }
)

func (MoveTo) isOp()       {}
func (LineTo) isOp()       {}
func (CubicTo) isOp()      {}
func (ClosePath) isOp()    {}
func (NewPath) isOp()      {}
func (Rect) isOp()         {}
func (Circle) isOp()       {}
func (RoundRect) isOp()    {}
func (SetColor) isOp()     {}
func (SetLineWidth) isOp() {}
func (SetCap) isOp()       {}
func (SetJoin) isOp()      {}
func (Save) isOp()         {}
func (Restore) isOp()      {}
func (Translate) isOp()    {}
func (Scale) isOp()        {}
func (Fill) isOp()         {}
func (FillEO) isOp()       {}
func (Stroke) isOp()       {}
func (Clear) isOp()        {}
