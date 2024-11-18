# ColourGen

ColourGen handles the input colour strings and transforms them
into go colour values.

All colours are used as 16 bit colours internally.
Therefore all non 16 bit colours are bit shifted by x bits to the 16 bit
value, e,g, a 8 bit red of 255 is shifted by 8 bits to a 16 bit value of 65280.
However a if a max alpha is used for a non 16 bit colour then the alpha
is given as the max 16bit value of 65535. This is to prevent the
transparency changing the intended values of the colours.

The following formats are used
with format - example - 16 bit value:

- `^#[A-Fa-f0-9]{6}$` - e.g. #FFFFFF - `color.NRGBA64{R:65280, G:65280, B:65280, A:65535}`
- `^#[A-Fa-f0-9]{3}$` - e.g. #FFF - `color.NRGBA64{R:61440, G:61440, B:61440, A:65535}`
- `^#[A-Fa-f0-9]{8}$` - e.g. #FFFFFFFF - `color.NRGBA64{R:65280, G:65280, B:65280, A:65535}`
- `^#[A-Fa-f0-9]{4}$` - e.g. #FFFF - `color.NRGBA64{R:61440, G:61440, B:61440, A:65535}`
- `^(rgba\()\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\)$` - rgba(255,255,255,255) - `color.NRGBA64{R:65280, G:65280, B:65280, A:65535}`
- `^(rgb\()\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\)$` - rgb(255,255,255) - `color.NRGBA64{R:65280, G:65280, B:65280, A:65535}`
- `^rgb12\(([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5])\)$` - rgb12(4095,4095,4095) - `color.NRGBA64{R:65520, G:65520, B:65520, A:65535}`
- `^rgba12\(([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5])\)$` - rgb12a(4095,4095,4095,4095) - `color.NRGBA64{R:65520, G:65520, B:65520, A:65535}`
