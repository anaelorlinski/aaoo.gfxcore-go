package gfxcore

import (
	"fmt"
	"strings"
)

// StrokeJoinMode type to specify how segments join when stroking.
type StrokeJoinMode uint8

// StrokeJoinMode constants determine how stroke segments bridge the gap at a join
const (
	Miter StrokeJoinMode = iota
	Round
	Bevel
)

func (s StrokeJoinMode) String() string {
	switch s {
	case Round:
		return "Round"
	case Bevel:
		return "Bevel"
	case Miter:
		return "Miter"
	default:
		return fmt.Sprintf("<unknown JoinMode %s>", string(s))
	}
}

// StrokeCapMode defines how to draw caps on the ends of lines
// when stroking.
type StrokeCapMode uint8

const (
	ButtCap StrokeCapMode = iota // default value
	RoundCap
	SquareCap
)

func (c StrokeCapMode) String() string {
	switch c {
	case ButtCap:
		return "ButtCap"
	case SquareCap:
		return "SquareCap"
	case RoundCap:
		return "RoundCap"
	default:
		return fmt.Sprintf("<unknown CapMode %s>", string(c))
	}
}

// StrokeOptions specifies advanced stroking options.
type StrokeOptions struct {
	LineCap  StrokeCapMode
	LineJoin StrokeJoinMode

	// MiterLimit is the miter cutoff value for `Miter`, `Arc`, `Miterclip` and `ArcClip` join modes
	MiterLimit Fl
}

// PaintOp specifies the graphic operation applied to the current path
type PaintOp uint8

const (
	Stroke PaintOp = 1 << iota
	FillEvenOdd
	FillNonZero // mutually exclusive with FillEvenOdd
)

func (op PaintOp) String() string {
	var chunks []string
	if op&Stroke != 0 {
		chunks = append(chunks, "stroke")
	}
	if op&FillEvenOdd != 0 {
		chunks = append(chunks, "fill even odd")
	}
	if op&FillNonZero != 0 {
		chunks = append(chunks, "fill non zero")
	}
	return strings.Join(chunks, " ")
}

type GradientKind struct {
	// Kind is either:
	// 	"solid": Colors is then a one element array and Positions and Coords are empty.
	// 	"linear": Coords is (x0, y0, x1, y1)
	// 			  coordinates of the starting and ending points.
	// 	"radial": Coords is (cx0, cy0, radius0, cx1, cy1, radius1)
	// 			  coordinates of the starting end ending circles
	Kind   string
	Coords [6]Fl
}

// Interpolation says how the colours between two stops of a gradient are
// computed.
type Interpolation uint8

const (
	// InterpolatePremultiplied averages the colours multiplied by their
	// alpha: a stop fading to transparent keeps its hue, whatever the
	// colour of the transparent stop. It is the way of CSS gradients.
	InterpolatePremultiplied Interpolation = iota
	// InterpolateStraight averages the colour and the alpha separately: a
	// fade from transparent white passes through pale tints. It is the
	// way of SVG gradients.
	InterpolateStraight
)

// GradientLayout is a gradient resolved to plain geometry: the renderer
// interpolates between the stops and needs nothing else.
type GradientLayout struct {
	// Positions is a list of floats in [0..1].
	// 0 at the starting point, 1 at the ending point.
	Positions []Fl
	Colors    []RGBA

	// Exponents, if non-nil, carries the interpolation exponent for the gap
	// that starts at each stop (parallel to Positions/Colors; the last
	// entry is unused). Value 0 means linear (the default). It lets the
	// renderer interpolate a curve natively (PDF Type 2 function exponent,
	// per-pixel pow) instead of the caller pre-sampling it into extra
	// stops. nil = all-linear.
	Exponents []Fl

	// Interpolation says in what terms two colours are mixed; Exponents
	// says how fast the mix goes from one to the other.
	Interpolation Interpolation

	GradientKind

	// ScaleY is used for elliptic radial gradients. 1 otherwise.
	ScaleY    Fl
	Repeating bool
}
