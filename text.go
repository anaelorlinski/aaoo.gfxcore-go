package gfxcore

import "github.com/anaelorlinski/aaoo.gfxcore-go/matrix"

// GID is a glyph index in a font.
type GID = uint32

// FontOrigin is a reference to a binary font file, either on disk or stored
// in memory.
type FontOrigin struct {
	// File is the filename or identifier of the font file.
	File string
	// Index is the face's index in a collection; 0 for a single font file.
	Index uint16
	// Instance is 1 + the variable-font instance index, or 0 for none.
	Instance uint16
}

// Font identifies the font of a [TextRun]. It is a handle: the code that
// lays text out and the renderer that draws it agree on the concrete type
// behind it, and nothing in between looks inside.
//
// Font are implemented by valid map keys.
type Font interface {
	Origin() FontOrigin
}

// TextDrawing exposes the positionned text glyphs to draw
// and the associated font, in a renderer independent manner.
type TextDrawing struct {
	// All the runs should be displayed sequentially
	Runs []TextRun

	FontSize, ScaleX Fl
	X, Y             Fl // origin
	Angle            Fl // (optional) rotation

	Text []rune
}

// Matrix return the transformation scaling the text by [FontSize],
// translating if to (X, Y)  and applying the [Angle] rotation
func (td TextDrawing) Matrix() matrix.Transform {
	mat := matrix.New(td.ScaleX, 0, 0, -1, td.X, td.Y)
	if td.Angle != 0 { // avoid useless multiplication if angle == 0
		mat.RightMultBy(matrix.Rotation(td.Angle))
	}
	return mat
}

// TextRun is a serie of glyphs with constant font.
type TextRun struct {
	Font   Font
	Glyphs []TextGlyph
}

// TextGlyph stores a glyph and it's position
type TextGlyph struct {
	Glyph GID

	Offset Fl // normalized by FontSize
	Rise   Fl

	XAdvance Fl // how much to move before drawing, used for emojis

	// TextDrawing.Text[TextOffset:TextOffset+TextLength]
	// gives the runes yielding the glyph.
	TextOffset, TextLength int
}
