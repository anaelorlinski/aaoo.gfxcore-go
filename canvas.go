package gfxcore

import "github.com/anaelorlinski/aaoo.gfxcore-go/matrix"

// Pattern is a paint that repeats a tile: it acts as a fill or stroke color,
// with arbitrary content.
//
// The tile is a rectangle of Width by Height, whose top left corner is the
// origin of the pattern. It is repeated every StepX and StepY: a step
// larger than the tile leaves an empty space between two tiles, and a step
// larger than what is painted shows the tile once.
type Pattern struct {
	// Tile is what is drawn in the tile, a canvas returned by
	// [Canvas.NewPatternTile]. What it draws out of the tile is not shown.
	Tile Canvas

	// Width and Height are the size of the tile.
	Width, Height Fl

	// StepX and StepY are the distances between two tiles; zero is the size
	// of the tile.
	StepX, StepY Fl

	// Transform maps the coordinates of the pattern to the ones it is
	// painted in.
	Transform matrix.Transform

	// Opacity is the opacity of what is painted, in [0, 1]; zero is opaque,
	// as one is.
	Opacity Fl
}

// GraphicState exposes the settings for a group of graphic
// operations.
type GraphicState interface {
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

	// SetColorPattern sets the paint of the fills, or of the strokes, to a
	// pattern: a tile repeated at fixed intervals, see [Pattern].
	SetColorPattern(pattern Pattern, stroke bool)

	// SetColorGradient sets the paint of the fills, or of the strokes, to a
	// linear or radial gradient. `mat` maps the coordinates of the gradient
	// to the ones it is painted in: the gradient is not bounded by a
	// rectangle, it paints whatever is filled or stroked with it.
	SetColorGradient(gradient GradientLayout, mat matrix.Transform, stroke bool)

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
// or as intermediate container (see NewGroup and NewPatternTile)
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

	// NewGroup creates a new drawing target, to be drawn on and then put on
	// this canvas as a whole by `DrawGroup`, or used as the mask of another
	// group. It is drawn in the coordinates of this canvas: it starts with
	// its current transformation.
	//
	// The rectangle is the area the group is known to draw in, in those
	// coordinates; a renderer may use it to bound what it allocates. A group
	// whose extent is not known when it is created is given it afterwards
	// with `SetBoundingBox`.
	NewGroup(x, y, width, height Fl) Canvas

	// DrawGroup puts a group created by `NewGroup` on this canvas, taken as
	// a whole: with an opacity, through a mask, with a blend mode, see
	// [Composite].
	//
	// The group is put where it was drawn. If the transformation of this
	// canvas has changed since the group was created, the group moves by as
	// much: a group can be drawn, then placed, and placed several times.
	DrawGroup(group Canvas, composite Composite)

	// NewPatternTile creates a new drawing target for the tile of a
	// [Pattern], of the given size. It is drawn in coordinates of its own,
	// whose origin is the top left corner of the tile, whatever the current
	// transformation of this canvas is.
	//
	// A tile that only holds one raster image is kept as that image: a
	// renderer draws it from its source, as [Canvas.DrawRasterImage] does.
	NewPatternTile(width, height Fl) Canvas

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
}
