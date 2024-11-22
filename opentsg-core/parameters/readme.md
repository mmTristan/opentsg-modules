# Parameters

Parameters contains parameters for widgets to use, where they may
be several input formats that can be used, which all can be converted to one value.

This is currently used for calculating internal offsets a#nd rotational angles.

The structs in this package are designed to be [embedded in the widget struct](https://gobyexample.com/struct-embedding)
to carry over the fields and methods, preserving the uniformity for these
generic parameters that are used across widgets.

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
