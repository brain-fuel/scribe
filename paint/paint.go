// Package paint provides paint sources for compositing masks. v1 has
// solid colors; gradients come later.
package paint

import (
	"image"
	"image/color"
)

// Solid returns an infinite uniform paint source of color c.
func Solid(c color.Color) image.Image { return image.NewUniform(c) }
