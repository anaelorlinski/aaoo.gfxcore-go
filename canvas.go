package gfxcore

import "github.com/anaelorlinski/aaoo.gfxcore-go/matrix"

// GraphicState exposes the settings for a group of graphic
// operations.
type GraphicState interface {
	// SetAlphaMask inteprets `mask` as an alpha mask which should be applied
	// to the painting done in the current graphic stack.
	SetAlphaMask(mask Canvas)

	// Establishes a new clip region
	// by intersecting the current clip region
	// with the current path as it would be filled by `Fill`
	// and according to the fill rule given in `evenOdd`.
	//
	// After `Clip`, the current path will be cleared (or closed).
	//
	// The current clip region affects all drawing operations
	// by effectively masking out any changes to the surface
	// that are outside the current clip region.
	//
	// Calling `Clip` can only make the clip region smaller,
	// never larger, but you can call it in the `OnNewStack` closure argument,
	// so that the original clip region is restored afterwards.
	Clip(evenOdd bool)

	// SetAlpha is the same as SetColorRgba, but only updates the alpha value.
	SetAlpha(alpha Fl, stroke bool)

	// Sets the color which will be used for any subsequent drawing operation.
	//
	// The color and alpha components are
	// floating point numbers in the range 0 to 1.
	// If the values passed in are outside that range, they will be clamped.
	// `stroke` controls whether stroking or filling operations are concerned.
	SetColorRgba(color RGBA, stroke bool)

	// SetColorPattern set the current paint color to the given pattern.
	// A pattern acts as a fill or stroke color, but permits complex textures.
	// It consists of a rectangle, fill with arbitrary content, which will be replicated
	// at fixed horizontal and vertical intervals to fill an area.
	// (contentWidth, contentHeight) define the size of the pattern content.
	// `mat` maps the pattern’s internal coordinate system to the one
	// in which it will painted.
	// `stroke` controls whether stroking or filling operations are concerned.
	SetColorPattern(pattern Canvas, contentWidth, contentHeight Fl, mat matrix.Transform, stroke bool)

	// SetBlendingMode sets the blending mode, which is a CSS blend mode
	// keyword ("multiply", "screen", ...), shared by CSS and SVG.
	SetBlendingMode(mode string)

	// Sets the current line width to be used by `Stroke`.
	// The line width value specifies the diameter of a pen
	// that is circular in user space,
	// (though device-space pen may be an ellipse in general
	// due to scaling / shear / rotation of the CTM).
	SetLineWidth(width Fl)

	// Sets the dash pattern to be used by `Stroke`.
	// A dash pattern is specified by dashes, a list of positive values.
	// Each value provides the length of alternate "on" and "off"
	// portions of the stroke.
	// `offset` specifies a non negative offset into the pattern
	// at which the stroke begins.
	//
	// Each "on" segment will have caps applied
	// as if the segment were a separate sub-path.
	// In particular, it is valid to use an "on" length of 0
	// with `RoundCap` or `SquareCap`
	// in order to distribute dots or squares along a path.
	//
	// If `dashes` is empty dashing is disabled.
	// If it is of length 1 a symmetric pattern is assumed
	// with alternating on and off portions of the size specified
	// by the single value.
	SetDash(dashes []Fl, offset Fl)

	// SetStrokeOptions sets additionnal options to be used when stroking
	// (in addition to SetLineWidth and SetDash)
	SetStrokeOptions(StrokeOptions)

	// GetTransform returns the current transformation matrix (CTM).
	GetTransform() matrix.Transform

	// Modifies the current transformation matrix (CTM)
	// by applying `mt` as an additional transformation.
	// The new transformation of user space takes place
	// after any existing transformation.
	Transform(mt matrix.Transform)

	// SetTextPaint adjusts how text shapes are rendered.
	SetTextPaint(op PaintOp)
}

// Canvas represents a 2D surface which is the target of graphic operations.
// It may be used as the final output (like a PDF page or the screen),
// or as intermediate container (see for instance DrawWithOpacity or SetAlphaMask)
//
// Canvas is what every renderer must implement. What only some renderers
// can do better than the generic way is not here: it is an optional
// interface, looked for on the canvas being drawn on. See docs/design.md.
type Canvas interface {
	// Returns the current canvas rectangle
	GetBoundingBox() (left, top, right, bottom Fl)

	// Updates the current canvas rectangle
	SetBoundingBox(left, top, right, bottom Fl)

	// OnNewStack save the current graphic state,
	// execute the given closure, and restore the state.
	OnNewStack(func())

	// State returns the current graphic state.
	State() GraphicState

	// NewGroup creates a new drawing target with the given
	// bounding box. It may be filled by graphic operations
	// before being passed to the `DrawWithOpacity`, `SetColorPattern`
	// and `SetAlphaMask` methods.
	NewGroup(x, y, width, height Fl) Canvas

	// DrawWithOpacity draw the given target to the main target, applying the given opacity (in [0,1]).
	DrawWithOpacity(opacity Fl, group Canvas)

	// Paint actually shows the current path on the target,
	// either stroking, filling or doing both, according to `op`.
	// The result of the operation depends on the current fill and
	// stroke settings.
	// After this call, the current path will be cleared.
	Paint(op PaintOp)

	// Adds a rectangle of the given size to the current path,
	// at position ``(x, y)`` in user-space coordinates.
	// (X,Y) coordinates are the top left corner of the rectangle.
	// Note that this method may be expressed using MoveTo and LineTo,
	// but may be implemented more efficiently.
	Rectangle(x Fl, y Fl, width Fl, height Fl)

	// Begin a new sub-path.
	// After this call the current point will be (x, y).
	MoveTo(x Fl, y Fl)

	// Adds a line to the path from the current point
	// to position (x, y) in user-space coordinates.
	// After this call the current point will be (x, y).
	// A current point must be defined before using this method.
	LineTo(x Fl, y Fl)

	// Add cubic Bézier curve to current path.
	// The curve shall extend to (x3, y3) using (x1, y1) and (x2,
	// y2) as the Bézier control points.
	CubicTo(x1, y1, x2, y2, x3, y3 Fl)

	// ClosePath add a straight line to the beginning of
	// the current sub-path (specified by MoveTo)
	// It is somewhat equivalent to adding a LineTo instruction,
	// but some backends may optimize the corner rendering, applying line join style.
	ClosePath()

	// DrawText draws the given text using the current fill color.
	// The rendering may be altered by a previous `SetTextPaint` call.
	DrawText(texts []TextDrawing)

	// DrawRasterImage draws the given image at the current point, with the given dimensions.
	// Typical format for image.Content are PNG, JPEG, GIF.
	DrawRasterImage(image RasterImage, width, height Fl)

	// DrawGradient draws the given gradient at the current point.
	// Solid gradient are already handled, meaning that only linear and radial
	// must be taken care of.
	// The gradient fills the rectangle (0, 0, width, height).
	DrawGradient(gradient GradientLayout, width, height Fl)
}
