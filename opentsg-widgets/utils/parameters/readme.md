# Parameters

Parameters contains parameters for widgets to use, where they may
be several input formats that can be used, which all can be converted to one value.

This is currently used for calculating internal offsets a#nd rotational angles.

The structs in this package are designed to be [embedded in the widget struct](https://gobyexample.com/struct-embedding)
to carry over the fields and methods, preserving the uniformity for these
generic parameters that are used across widgets.

## Colour

Parameters offers several string formats to generate colour options
across openTSG.

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

## Clockwise Rotation

Clockwise rotation handle the rotational angle of
the widget, all values are converted to radians to
be used in the widget.

Clockwise rotation contains the following field names:

- `"cwRotation"`

These fields can be the following parameterisations,
where they would satisfy the regexes, if they are a string.

- Radian parameter: This is a string that matches the regexes of `π\*(\d){1,4}/{1}(\d){1,4}$` or
`^π\*(\d){1,4}$`. Examples include `π*1`, `π*23/47` and `π*3/4`
- Degree parameter: This is a number parameter that is a float between 0 and 360.
Examples include `270.0321`, `173.356` and `90`.

## Offset

Offset handles the offset in the internal widget canvas,
all values are converted to the exact pixel value.

Offset contains the following field and nested field names:

- `"offset"`
  - `"x"`
  - `"y"`

These fields utilise the following units, the units
used in the x and y fields do not need to match.

- Pixel parameter: This is a string that matches the regex of `^-{0,1}\d{1,}px$`.
Examples include `"20px"`, `"10000px"` and `"-55px"`
- Percentage string parameter: This is a string that matches the regex of
`-{0,1}\d{0,2}\.{1}\d{0,}%$|^-{0,1}\d{0,2}%$|^-{0,1}(100)%$`
Examples include `"20%"`, `"-34.34%"` and `"99.9997%"`.
- Percentage Number parameter: This is a number parameter that is a float between 0 and 360.
Examples include `20`, `-34.34` and `99.9997`.
