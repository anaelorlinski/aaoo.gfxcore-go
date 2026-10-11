package gfxcore

import "image/color"

// RGBA is a color in sRGB, components and alpha in [0, 1],
// not premultiplied.
type RGBA struct {
	R, G, B, A float32
}

func (c RGBA) Unpack() (r, g, b, a float32) {
	return c.R, c.G, c.B, c.A
}

// IsNone reports the zero value, used for "no color".
func (c RGBA) IsNone() bool {
	return c == RGBA{}
}

func clamp(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

var _ color.Color = RGBA{}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the color. Each value ranges within [0, 0xffff], but is represented
// by a uint32 so that multiplying by a blend factor up to 0xffff will not
// overflow.
//
// An alpha-premultiplied color component c has been scaled by alpha (a),
// so has valid values 0 <= c <= a.
func (c RGBA) RGBA() (r, g, b, a uint32) {
	c.R, c.G, c.B, c.A = clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A)
	// rounded to the nearest: cutting the fraction off would make every
	// colour a little darker than it is
	sa := 0xFFFF * c.A
	return uint32(c.R*sa + 0.5), uint32(c.G*sa + 0.5), uint32(c.B*sa + 0.5), uint32(sa + 0.5)
}
