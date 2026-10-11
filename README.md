# aaoo.gfxcore-go

The drawing vocabulary shared by the libraries that produce drawing
(`aaoo.svg-go`, `aaoo.webengine-go`) and the code that renders it.

It depends on the standard library only: no CSS engine, no text engine, no
renderer.

## Content

- `gfxcore`:
  - `Canvas` and `GraphicState`, the interface a renderer implements;
  - the types its methods take: `RGBA`, `PaintOp`, stroke options,
    `GradientLayout`, `RasterImage`, `TextDrawing` and `Font`;
  - the float type `Fl`.
- `gfxcore/matrix`: 2D affine transforms.

See [docs/design.md](docs/design.md) for the design: what is required of a
renderer, how optional capabilities work, and what is planned.

## Origin

The code comes from the `backend` and `matrix` packages of
[benoitkugler/webrender](https://github.com/benoitkugler/webrender),
a Go port of WeasyPrint. See `LICENSE`.
