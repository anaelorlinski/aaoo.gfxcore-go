package gfxcore

// RoundedBox is a rectangle with (possibly zero) elliptical corner radii,
// expressed in device coordinates. Each corner is (rx, ry).
type RoundedBox struct {
	X, Y, Width, Height                        Fl
	TopLeft, TopRight, BottomRight, BottomLeft [2]Fl
}
