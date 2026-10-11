package gfxcore

// BlendMode is how a group is mixed with what is under it: the blend modes
// of CSS Compositing and of PDF, which are the same.
type BlendMode uint8

const (
	BlendNormal BlendMode = iota
	BlendMultiply
	BlendScreen
	BlendOverlay
	BlendDarken
	BlendLighten
	BlendColorDodge
	BlendColorBurn
	BlendHardLight
	BlendSoftLight
	BlendDifference
	BlendExclusion
	BlendHue
	BlendSaturation
	BlendColor
	BlendLuminosity
)

var blendModeNames = [...]string{
	"normal", "multiply", "screen", "overlay", "darken", "lighten",
	"color-dodge", "color-burn", "hard-light", "soft-light", "difference",
	"exclusion", "hue", "saturation", "color", "luminosity",
}

// String returns the CSS keyword of the mode.
func (b BlendMode) String() string {
	if int(b) < len(blendModeNames) {
		return blendModeNames[b]
	}
	return "normal"
}

// ParseBlendMode returns the mode of a CSS keyword, and false for a keyword
// that is not one.
func ParseBlendMode(keyword string) (BlendMode, bool) {
	for i, name := range blendModeNames {
		if name == keyword {
			return BlendMode(i), true
		}
	}
	return BlendNormal, false
}

// MaskMode says what of a mask hides what it masks.
type MaskMode uint8

const (
	// MaskAlpha masks by the opacity of the mask.
	MaskAlpha MaskMode = iota
	// MaskLuminance masks by the luminance of the mask, times its opacity:
	// white shows, black and what is empty hide. It is the mask of SVG.
	MaskLuminance
)

// Composite says how a group is put on a canvas by [Canvas.DrawGroup]: taken
// as a whole, as if it were one image. Its zero value puts the group as it
// is.
//
// Everything that applies to an element as a whole is said here, at once: an
// element that is half transparent, masked and blended is one group.
type Composite struct {
	// Opacity is the opacity of the group, in [0, 1]; zero is opaque, as one
	// is. A group that is not to be seen is not drawn.
	Opacity Fl

	// Mask, when it is not nil, shows the group where the mask is opaque,
	// or bright for [MaskLuminance], and hides it where the mask is empty.
	// It is a group of the same canvas, see [Canvas.NewGroup], drawn in the
	// same coordinates as the group it masks.
	Mask     Canvas
	MaskMode MaskMode

	// Blend is how the group is mixed with what is under it.
	Blend BlendMode
}
