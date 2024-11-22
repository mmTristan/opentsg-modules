# Gridgen

Gridgen handles the grid coordinate system of the test card.
It Scales the rows and columns to generate grids that fit the size of the canvas,
and then draws the grid lines on the base image.

It is the first step in making the base test signal, as once the
coordinates are made then the widgets can be made.
As part of the coordinates system it generates any geometry from TSIGs.

It also has the art key functionality, to allow for different
backgrounds to be used, where the background is to be preserved.

## The location system

The coordinates of an area a widget covers can be defined
in several formats, but they all use the same fields.
The (0,0) coordinate is located in the top left of the testcard,
and the widget positions are generated with the following json layout.

```json
"location": {
        "box": {
            "x": 0,
            "y": 0
        }
    }
```

The default height and width of a widget is one grid unit,
if you want to extend it you can with 2 options (in order of precedence):

1. Set the x2,y2 fields
2. Set the height and width

These fields can be mixed and matched.

The x2 and y2 fields denote the bottom right coordinate of the widget
and can be created with the following json.

```json
"location": {
        "box": {
            "x": 1,
            "y": 1,
            "x2":5,
            "y2":12
        }
    }
```

The height and width are the height and width of the widget. The height is the distance down to
the bottom of the widget, and the width is the distance to the right of the widget.
They can be created with the following json, which creates the same result as the previous demo
for x2 and y2.

```json
"location": {
        "box": {
            "x": 1,
            "y": 1,
            "width":4,
            "height":11
        }
    }
```

### Distance units

There are several units that can be used to call the coordinates.

Grid coordinates, these are the x,y values of the grid called in the canvas widget.
Called with the following style values:

- `1`
- `"1"`

Percentage units, these are the percentage of the total dimension (height or width).
They can be called like so.

- `"20%"`

Pixel units, these are the absolute pixel values on the test card.
and are called as so.

- `"500px"`

### border radius

A widget with rounded corners can be created with the
`"border-radius"` field which is implemented like so.

```json
"location": {
        "box": {
            "x": 1,
            "y": 1,
            "width":4,
            "height":11,
            "border-radius" : "20%"
        }
    }
```

This uses all the same units as the coordinates but with the percentage
being the height and width of the widget (which ever is smallest), instead of the whole
testcard.

## The legacy coordinates systems

There are multiple ways of defining the coordinates of an area a widget covers.
They are listed below to help you decide which method you prefer to use.

### Spreadsheet style

The spread sheet style coordinates are bound by the grid system of
openTSG. The letters define the X axis starting at A, and the numbers
the Y axis.
A1 is the start coordinate, and is the top left most grid on the test signal.
X values increase from left to right, and y values increase from top to bottom.

All values must satisfy the following regexes:

- `^[a-zA-Z]{1,3}[0-9]{1,3}$"` - For a single grid
- `^[a-zA-Z]{1,3}[0-9]{1,3}:[a-zA-Z]{1,3}[0-9]{1,3}$"` - for a section of grids,
these must go from top left most grid to the bottom right most grid.

### Alias

Alias is a string for a known location. The aliases are preserved
throughout the run of OpenTSG and redeclaring an alias will overwrite it.
The alias is only checked if it used in the `location` field.

All alias satisfy the regex of `^[\w\W]{1,30}$`

### Pixel Perfect

You can place widgets exactly where you want them based on their pixel position.
This does change the grid system of OpenTSG, rather it ignores it.
This method does not scale well if you want to make the test pattern for
multiple image definitions, as there pixel values will remain the same and not scale.

All pixel coordinates must satisfy the following regex `^\([0-9]{1,5},[0-9]{1,5}\)-\([0-9]{1,5},[0-9]{1,5}\)$`,
which is pixel inclusive. e.g. `"(0,0)-(1000,1000)"` would cover every pixel from
(0,0), (0,1000), (1000,0) and 1000,(1000).

### Row Column

The RowColumn style coordinates are bound by the grid system of
openTSG. The X axis is defined by the row, starting at C1, and
the Y axis is defined by the row, starting at R1.
R1C1 is the start coordinate, and is the top left most grid on the test signal.
X values increase from left to right, and y values increase from top to bottom.

All values must satisfy the following regexes:

- `^[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1})$` - For a single box
- `^[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1}):[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1})$` - for
a section of boxes,
these must go from top left most grid to the bottom right most grid.

## Art Key

An image can be used as a background image for a test signal.
This image can then have cutaways to empty images that can be keyed as areas to fill in
and masked.

## TSIG (Test Signal Input Geometry)

TSIGs are used for generating test patterns to fit 3d shapes.
They can be included within the canvas options json, with
the following code.

```json
{
    "type": "builtin.canvasoptions",
    ...
    "geometry":"./path/to/TSIG.json"
}

```

The `"geometry"` field specifies the TSIG file, no other input is required
to include TSIGs.

### How do TSIGS work

The geometry is the first thing openTSG checks when generating the base test pattern.
If a TSIG is present, the TSIG is unwrapped and sets the layout of the test pattern
that can be written to. The OpenTSG grid system still overlays this base image.

OpenTSG then runs normally, except that any pixels that aren't within the TSIG boundaries
are not filled in.

After OpenTSG has run the image is saved as normal, unless carving is required.
If carving is required then the image is then carved up into the smaller images
and each carved image is saved individually,
as {filename}{carve}.ext. The flat image is also saved
Find out more about carving with TSIGS [here](#carving)

### What is in a TSIG

A TSIG is a json file, documenting the flat layout, any
carving required and individual tiles of the object.

It has the following layout:

```json
{
    "Tile layout": [
        {
            "Name": "A000",
            "Tags": [],
            "Layout": {
                "Carve": {
                    "Destination": "C1",
                    "X": 0,
                    "Y": 0
                },
                "Flat": {
                    "X": 0,
                    "Y": 0
                },
                "XY": {
                    "X": 10,
                    "Y": 10
                }
            }
        }
    ],
    "Dimensions": {
        "Flat": {
            "X0": 0,
            "Y0": 0,
            "X1": 30,
            "Y1": 30
        }
    },
    "Carve": {
        "C1": {
            "X0": 0,
            "Y0": 0,
            "X1": 30,
            "Y1": 30
        }
    }
}
```

The `"Tile layout"` field contains a data array for each tile face. It contains

- `"Name"` - the name of that polygon, often used for labelling. - OPTIONAL
- `"Tags"` - any tags associated with that polygon, see [here](#tsig-tags) for more information about tags. - OPTIONAL
- `"Layout"` - the XY coordinate layout of the face. It has the following sub fields:
  - `"Carve"` - the name of the carve destination, and the XY coordinates - OPTIONAL
  - `""Flat"` - The initial XY coordinates of the object in its flat layout. - REQUIRED
  - `"XY"` - The height and width of the face. - REQUIRED

For carve and flat only the initial XY coordinates are required,
as the XY width and height gives the size of the tile.

An example `"Tile layout"` is given below.

```json
{
            "Name": "A000",
            "Tags": [],
            "Layout": {
                "Carve": {
                    "Destination": "C1",
                    "X": 0,
                    "Y": 0
                },
                "Flat": {
                    "X": 0,
                    "Y": 0
                },
                "XY": {
                    "X": 10,
                    "Y": 10
                }
            }
        }
```

The `"Dimensions"` contains the fields for the flat object dimensions.
The example below has a size of 30 by 30 pixels, X0,Y0 are the minimum coordinates
and X1,Y1 represent the maximum coordinates.

```json
"Dimensions": {
        "Flat": {
            "X0": 0,
            "Y0": 0,
            "X1": 30,
            "Y1": 30
        }
    }
```

The `"Carve"` field is a map of carve name and its size, to allow
different carve destinations to have different dimensions.

e.g.

```json
 "Carve": {
        "Carve1": {
            "X0": 0,
            "Y0": 0,
            "X1": 30,
            "Y1": 30
        },
        "Carve2": {
            "X0": 0,
            "Y0": 0,
            "X1": 550,
            "Y1": 2000
        }
    }
```

The example above has two carve destinations, carve1 and carve2 of varying sizes.
Where X0,Y0 are the minimum coordinates
and X1,Y1 represent the maximum coordinates.

Please note coordinates, start at (0,0) in the top left of the canvas
and go up to (max_X,max_Y) in the bottom right.

### TSIG tags

Each face has a tag field, for adding additional information about the face.
Some of these tags are used by the widgets, the ones currently in use are:

- `"neighbour:example"` - this is used to signify what the polygon's neighbours are. It is currently used by the four colour widget

You can add your own custom tags to describe extra info about the tile, as long
as it does not conflict with existing fields.

### Carving

Carving is required when several UV maps are required to put a picture on an object.
It is used for very large or complex objects,
where several inputs are required to light up different
sections of the display.

To implement this in a TSIG the carve field needs to be declared
with the name of the carved image and its dimensions like so.

In the example below the carve has one file called C1,
it has dimensions of a width of 300 and a height of 300.

```json
"Carve": {
        "C1": {
            "X0": 0,
            "Y0": 0,
            "X1": 300,
            "Y1": 300
        }
    }
```

Then each tile to be carved requires a `"Carve"` field with the `"Destination"`
of the carve matching the name of a carve location,
and the XY coordinate within the carved image e.g.

In the example below, the tile has a flat layout at (0,0) and
when it is carved it is assigned C1 at (10,10)

```json
 {
            "Name": "A000",
            "Tags": [],
            "Layout": {
                "Carve": {
                    "Destination": "C1",
                    "X": 10,
                    "Y": 10
                },
                "Flat": {
                    "X": 0,
                    "Y": 0
                },
                "XY": {
                    "X": 10,
                    "Y": 10
                }
            }
        }
```

### TSIGS and widgets

Some widgets are designed for use with TSIG as part of the input, for
example the four colour widget.
When these widgets are run without a TSIG, the grid coordinate system
is used as the input geometry instead.

The complete list of TSIG widgets is given below.

- [Four colour](../../../opentsg-widgets/_docs/fourcolour/doc.md)

### Building a TSIG

These are the steps we undertook to build a simple TSIG
for a house in the [TSIG demo](https://github.com/mrmxf/opentsg-node/blob/main/READMETPIG.md).

We started with making an obj of a cube.
We ignored the normals and only focused on geometric vertices and texture
coordinates of the tiles.

The first step was making the coordinates of the cube,
so that it unwrapped as shown in the readme. This did not work first time
as it turns out blender has to be told what axis is up, when importing objs,
as there is no standard.
Blenders default is y up, but the obj was designed for z to be the up axis.

This leads to many minutes pondering if the axis lied to me in the obj format,
but it was only blender making everything wonky.

Next was getting the uv map correct, so when it was unfolded the textures lined up.
This meant that tiles textures aren't reversed or upside down, due to the
u,v coordinates not lining up.

This meant going through the UV map ensuring coordinates
for neighbouring vertices are the same. And ensuring the
sizes of the map scales correctly to the physical sie of
the display, to prevent squeezing or stretching of the image.
e.g. left panel top right matches the top left of the next panel.

Once the obj then had a correct UV map and the tiles
all lined up as intended we moved onto the TSIG design.

The TSIG transferred the UV map to json,
gave a scale to set the base image, we took squares
to be 1080x1080 pixels. For example a UV coordinate
of 0.25,0.3333 then transferred to a TSIG coordinate
of 1080,1080, for the
unwrapped cube having a length of 4320 and a height
of pixels.

Each tiles uv map was then scaled to give pixel precise
values. No carving was added as the unwrapped image
was the uv map. If the uv map differs from the flat then carving
would have been added.
