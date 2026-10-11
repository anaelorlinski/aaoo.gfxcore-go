// Package gfxcore holds the drawing vocabulary shared by the libraries that
// produce drawing (the SVG library, webengine) and the code that renders it.
//
// It has no dependency: no CSS engine, no text engine, no renderer.
package gfxcore

// Fl is the float type used for all coordinates and lengths.
type Fl = float32
