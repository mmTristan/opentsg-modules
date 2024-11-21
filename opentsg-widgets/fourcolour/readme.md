# Four Colour

Is the [four colour theorem](https://en.wikipedia.org/wiki/Four_color_theorem)
designed for Test Signal Input Geometry (TSIG). Where four
colours are the maximum number of colours to
fill a map, without any neighbours having the same colour.
For large complicated TSIGs where many tiles overlap, it
is useful to see the layout of the tiles, to help fix
overlap or uv map issues.
This widget allows 4+ colours to be used
in the interest of computational speed for large objects,
with many tiles to find.

Any TSIG input needs to have the neighbours defined
in the `"Tags"` field of the TSIG. These are defined
with the `"neighbour:"` prefix within the tag.
Like so

```json
"Tags": ["neighbour:A1","neighbour:A2"]
```

This is because four colour does not have the means to calculate
the neighbours of each tile itself.

It has the following required field:

- `colors` - a list of TSG colors, it must be an array
of four strings or more. The algorithm will use the fewest amount of colors required.
The colours follow the [TSG colour formats](../utils/parameters/readme.md#colour)

```json
{
    "type" :  "builtin.fourcolour",
    "colors": [
        "#FF0000",
        "#00FF00",
        "#0000FF",
        "#FFFF00"
    ],
    "grid": {
      "location": "a1",
      "alias" : "A demo Alias"
    }
}
```
