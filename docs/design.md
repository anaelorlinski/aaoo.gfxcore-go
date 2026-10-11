# Design notes

## What this module is

`gfxcore` is the contract between two kinds of code:

- **producers**, which decide what to draw: the SVG library (`aaoo.svg-go`)
  and `aaoo.webengine-go` (HTML/CSS);
- **renderers**, which draw it: `aaoo.drawengine-go` today, anything else
  tomorrow.

Producers import `gfxcore` and nothing of a renderer. A renderer implements
`gfxcore.Canvas` and nothing of a producer. `aaoo.drawengine-go` does it in
its `gfx` package, beside its own canvas, rasterizer and PDF writer, which
know nothing of `gfxcore`.

The module depends on the standard library only.

## The two levels of the interface

### Required: `Canvas` and `GraphicState`

What every renderer must do, and what the generic drawing of SVG and CSS is
written with:

| group | methods |
|---|---|
| path | `MoveTo` `LineTo` `CubicTo` `ClosePath` `Rectangle` |
| painting | `Paint` `DrawGradient` `DrawRasterImage` `DrawText` |
| state | `OnNewStack` `State` and, on the state: `Transform` `GetTransform` `Clip` `SetColorRgba` `SetAlpha` `SetLineWidth` `SetDash` `SetStrokeOptions` `SetBlendingMode` `SetTextPaint` |
| layers | `NewGroup` `DrawWithOpacity`, and on the state: `SetAlphaMask` `SetColorPattern` |
| extent | `GetBoundingBox` `SetBoundingBox` |

Layers are required: opacity, masks and patterns cannot be drawn without
them, so a target that cannot hold a group cannot be a `Canvas`.

Text is required too, because text is drawn into groups like everything
else. A `Font` is a handle: the code that lays text out and the renderer
agree on the concrete type behind it, and `gfxcore` does not look inside.
A renderer with no text support implements `DrawText` as a no-op.

### Optional: capabilities

What only some renderers can do better than the generic way is **not** in
`Canvas`. It is a small interface of its own, which a canvas may also
implement. The producer looks for it on the canvas it is drawing on:

```go
if dst, ok := canvas.(backend.BoxShadowDrawer); ok {
	dst.DrawBoxShadow(...)
}
```

Rules:

1. **Every capability has a defined result when it is absent**: either a
   generic drawing through `Canvas`, or the effect is not drawn. The
   interface documents which.
2. **The lookup is on the current canvas, groups included.** A renderer
   implementing a capability must do so on every canvas it hands out from
   `NewGroup`.
3. **A capability lives with its producer** unless two producers need it.
   The three existing ones are CSS only and are in
   `aaoo.webengine-go/backend`: `BoxShadowDrawer`, `TextShadowDrawer`,
   `GradientFillPather`.

This is how a renderer gets to be better than the common denominator
without the others paying for it.

## Planned capabilities: passing originals through

Not implemented. They are recorded here because they constrain how the API
may evolve: nothing added to `gfxcore` or to the producers should make them
impossible.

The common idea: a producer often holds something richer than the drawing
calls it turns it into. A renderer that can use the original should get it,
instead of a lossy or heavier re-expression.

### Bitmaps

Already the case, and to be kept: `RasterImage` carries the **encoded**
image (`Content`, `MimeType`), not pixels, with an `ID` identifying the
content and the `Rendering` hint.

- A PDF renderer can embed a JPEG as it is, with no decode and re-encode.
- A raster renderer decodes once per `ID`.

Constraints to keep:

- producers never decode an image to hand pixels over;
- `ID` identifies the content, so that a renderer can share one copy;
- `Content` is a reader, read at most once per `ID`; if a renderer needs
  the bytes twice it keeps them.

### SVG

When a page contains an SVG image and the renderer writes SVG (or a format
that can embed SVG), the original document can be written out unchanged
instead of being replayed as paths.

Sketch:

```go
// VectorImageDrawer is implemented by a canvas able to take an SVG
// document as it is. It reports false when it did not, and the image is
// then drawn through Canvas.
type VectorImageDrawer interface {
	DrawSVG(source []byte, baseURL string, width, height Fl) bool
}
```

Where it plugs in: the host's image wrapper (`images.SVGImage.Draw` in
webengine), before it calls `svg.SVGImage.Draw`. The SVG library itself
only needs to keep the source reachable.

Constraints to keep:

- the SVG library must be able to give back the **original bytes** and base
  URL of a document it parsed (it keeps neither today);
- passthrough is only exact for a self-contained SVG *image*. An inline
  `<svg>` inherits `currentColor` and the device of its page; external
  resources and fonts follow the host's policy. The wrapper decides, and
  falls back to drawing;
- the capability must also work on a group, since images are drawn into
  pattern and opacity groups.

### Node data, for tagged PDF

A tagged PDF needs to know which drawing belongs to which element (see
`docs/tagged-pdf.md` in the compositor). The structure tree itself does not
have to travel through drawing; only a small marker does.

Sketch:

```go
// NodeMarker is implemented by a canvas able to attach what is drawn to a
// node of the document. Calls nest, and are balanced within one
// OnNewStack.
type NodeMarker interface {
	BeginNode(id NodeID)
	EndNode()
}
```

Constraints to keep:

- drawing of one element stays bracketable: producers draw an element's
  content between two points of their code, not interleaved with others;
- text keeps its source: `TextDrawing.Text` and the per glyph
  `TextOffset` / `TextLength` are what a renderer needs for real text,
  `/ActualText` and reading order;
- links and anchors, today on webengine's `Page`, are the same kind of
  data and would follow the same route.

## Open points

- `SetBlendingMode` takes a CSS keyword as a string; a typed enumeration
  would be safer.
- `GradientKind.Kind` is a string too.
- `GradientFillPather` is stateful ("applies to the next `DrawGradient`").
  A fill area passed with the gradient would be simpler; it has one caller.
- `SetColorRgba` / `SetAlpha` select fill or stroke with a boolean.
- `GetBoundingBox` / `SetBoundingBox` exist for groups whose extent is
  known only after drawing (SVG text) and for unbounded backgrounds.
