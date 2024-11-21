# Geometry Text

Geometry text prints the name of a tile of the Test Signal Input Geometry (TSIG). 
Where the text is squeezed to fill the tile.

It has the following required field:

- `textColor` - a string of TSG colors.
This follows the [TSG colour formats](../utils/parameters/readme.md#colour)

```json
{
    "props": {
        "type": "builtin.geometrytext",
        "location": {
            "box": {
                "x": 0,
                "y": 0
            }
        }
    },
    "textColor":"#000000"
}
```
