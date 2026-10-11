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
| painting | `Paint` `DrawRasterImage` `DrawText` |
| state | `OnNewStack` `State` and, on the state: `Transform` `GetTransform` `Clip` `SetLineWidth` `SetDash` `SetStrokeOptions` `SetTextPaint` |
| paint | on the state, for the fills or for the strokes: `SetColorRgba` `SetAlpha` `SetColorGradient` `SetColorPattern`, and `NewPatternTile` for the tile of a pattern |
| groups | `NewGroup` `DrawGroup`, with a `Composite`: opacity, mask, blend mode |
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
   `NewGroup` and `NewPatternTile`.
3. **A capability lives with its producer** unless two producers need it.
   The two existing ones are CSS only and are in
   `aaoo.webengine-go/backend`: `BoxShadowDrawer`, `TextShadowDrawer`.

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

## Colour

`RGBA` is a colour in sRGB: four floats, not premultiplied. It is what a
producer computed, and it reaches a renderer unchanged: the draw engine
keeps it in its paints and its gradient stops, writes it as it is in a PDF,
and rounds it once, to pixels, when it rasterizes. A transparent stop keeps
its colour, which a premultiplied type cannot hold.

A gradient says how two stops are mixed (`Interpolation`), because the
formats disagree: CSS mixes premultiplied colours, SVG the colours and the
alpha separately.

What is not there yet: a colour that names its space. hsl, lab, oklch and
the like are ways of writing or of mixing colours; the CSS engine converts
them, and a renderer only needs to know of them as the space a gradient is
mixed in. CMYK, grey, wide gamut RGB and ICC profiles are spaces a device
can be told to use: for those the colour type itself has to carry a space,
which is a study of its own (a PDF for print is the case that needs it).

## Paints

A fill or a stroke is painted with a colour, a gradient or a pattern, and
the three are set the same way, on the state. A gradient and a pattern are
not drawings: they have no rectangle of their own, they paint whatever path
or text is filled or stroked with them, through a matrix that places them.

A pattern is a tile repeated at a step, `Pattern`. Its tile is drawn on a
canvas of its own, `NewPatternTile`, in coordinates that start at the corner
of the tile. A step larger than the tile leaves a space between two tiles,
and one larger than what is painted shows the tile once: this is how the
`space` and `no-repeat` of a CSS background are said, with nothing special
for them. A tile that holds one raster image is kept as that image, so that
a renderer draws it from its source.

## Groups

Whatever applies to an element as a whole goes through one call. The
producer draws the element in a group, `NewGroup`, then puts the group on
the canvas with `DrawGroup` and a `Composite` that says how: its opacity,
its mask, its blend mode. An element that is half transparent, masked and
blended is one group, so one offscreen image for a rasterizer and one form
for a PDF.

There is no state for this. A mask or a blend mode set "from now on" would
have to be undone by the renderer at the end of a stack the producer does
not name; the group is the thing the effect applies to, and the producer
holds it.

Rules:

- **A group is drawn in the coordinates of the canvas it was created on**,
  as they were when it was created.
- **A group is put where it was drawn**, moved by what the transformation
  of the canvas has changed since. A producer that does not touch the
  transformation gets the group back in place; one that does places it.
- **A group may be put several times.** A renderer may then keep it once: a
  PDF form drawn at several places. Nothing does this yet (the PDF writer
  writes the form each time), but nothing prevents it.
- **A mask is a group** of the same canvas, drawn in the same coordinates
  as the group it masks, and moved with it.
- **A group that is not to be seen is not drawn.** The zero value of
  `Composite` is "as it is", so an opacity of zero cannot mean hidden.

What this shape is meant to carry later, without another change of the
interface:

| later | where it goes |
|---|---|
| filters of SVG and CSS (blur, colour matrix, drop shadow) | a field of `Composite`; a PDF has no filters, so a filtered group is an image there |
| `mix-blend-mode`, `isolation`, `mask-image` of CSS | the fields that exist; the web engine only has to fill them |
| a clip that is not a rectangle on a bitmap, an anti-aliased clip | a mask on a group |
| shadows as "the group, blurred and moved" | the filter field; today they are drawn from their geometry, which stays vector in a PDF |
| an SVG `<use>`, a symbol, a header repeated on every page | one group put several times |
| the bounds of a group | the rectangle of `NewGroup`, which no renderer reads yet: it would bound an offscreen image and be the box of a PDF form |
| structure for a tagged PDF | not here: see the node markers above. A marker is a record in what is drawn, so a group carries the markers of what was drawn in it. A group put several times is the case to watch: its content would belong to several nodes, and PDF wants each piece of content in one |

## Open points

- `GradientKind.Kind` is a string too.
- `SetColorRgba` / `SetAlpha` select fill or stroke with a boolean.
- `GetBoundingBox` / `SetBoundingBox` exist for groups whose extent is
  known only after drawing (SVG text) and for unbounded backgrounds.
