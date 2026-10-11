package gfxcore

import "io"

// RasterImage is an image to be included in the ouput.
type RasterImage struct {
	Content  io.Reader
	MimeType string

	// Rendering is the CSS property for this image.
	Rendering string

	// ID is a unique identifier which permits caching
	// image content when possible.
	ID int
}
