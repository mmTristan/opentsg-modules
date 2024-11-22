# tsg-widgets

`opentsg-widgets` is the library of builtin widgets,
for the [Open Test Signal Generator](https://opentsg.studio/).
Feel free to use any as a blueprint to make your own widgets,
for use with openTSG.

## Examples

Examples of the json for each type can be found at the `exampleJson` folder.
These examples contain all the fields unique to the widget. The name of the folder
is the same as the `"type"` field that would be declared for that widget. examples that have nested
folders such as builtin.ebu3373/bars, have their widget names including the `/`, so
`builtin.ebu3373/bars` would be the widget type.

### Legacy version

The positional fields of `"grid"` are expected to be in **every** widget input file. And have the
layout as shown below. They may not be found in every/any example json.
The widget type is also required. Each widget has a unique `"type"`,
so OpenTSG can identify and use the widget.

```json
"type" : "builtin.example",
"grid": {
    "location": "a1:b2",
    "alias" : "A demo Alias"
}

```

### Handler version

All openTSG widgets properties are stored in the `"props"` field. This props field
is not passed to the widget directly and does not need to accounted for in the schema.
The positional fields of `"location"` are expected to be in **every** widget input file. And have the
layout as shown below. See the grid [documentation](../opentsg-core/gridgen/readme.md#the-location-system) for more information
The widget type is also required. Each widget has a unique `"type"`,
so OpenTSG can identify and handle the widget.

```json
"props"{
    "type" : "builtin.example",
    "location": {
        "alias" : "A demo Alias",
        "box": {
            "x": 0,
            "y": 0
        }
    }
}

```

## Widget Properties

This section contains the properties of the widgets.
This contains the design behind the widget, the fields
and contents it uses. And an example JSON

- [AddImage](./addimage/readme.md)
- [Ebu3373](./ebu3373/readme.md)
- [Fourcolour](./fourcolour/readme.md)
- [FrameCount](./framecount/readme.md)
- [GeometryText](./geometryText/readme.md)
- [Gradients](./gradients/readme.md)
- [Noise](./noise/readme.md)
- [QrGen](./qrgen/readme.md)
- [TextBox](./textbox/readme.md)
- [ZonePlate](./zoneplate/readme.md)

## Notes for developers

There are several stages for developing a widget to use in OpenTSG.

The first is the configuration of the widget. It
requires the following items:

- The json schema of the object, preferably embedded in the code.
(one less file to track)
- The functions to match the widget handler interface.
- Most importantly an idea of what the widget is there to test for and
achieve.

The widget handler Generator interface

```go
// The handler bytes are parsed into an object that runs the Handle method.
type Handler interface {
    Handle(Response, *Request)
}
```

An example widget would look like this.
This example is a simple widget that fills in a canvas as a solid green.

```go

import (
    _ "embed"

    "github.com/mrmxf/opentsg-modules/opentsg-core/colour"
    "github.com/mrmxf/opentsg-modules/opentsg-core/config"
)


type Config struct {
    // fill out the fields here
}

// the example schema (this is not a real file)
//go:embed jsonschema/exampleSchema.json
var Schema []byte

const (
    WidgetType = "builtin.example"
)


// It just fills in the canvas as green (for this demo)
func (c Config) Handle(resp tsg.Response, req *tsg.Request) {

    draw.Draw(resp.BaseImage(), resp.BaseImage().Bounds(),  &image.Uniform{color.RGBA{G: 0xff, A: 0xff}}, image.Point{}, draw.Over)
    

    resp.Write(tsg.WidgetSuccess, "success")
}
```

Which can be added to tsg by calling the following code,
before the TSG object is run. Or if you want to add it to the standard library of
openTSG, then add it to line 41 of `/opentsg-widgets/widget.go`

The below is a demo of adding an external widget
without changing the opentsg library.

```golang

    // Set up the openTSG engine
    otsg, configErr := tsg.BuildOpenTSG(commandInputs, *profile, *debug, &tsg.RunnerConfiguration{RunnerCount: 1, ProfilerEnabled: true}, myFlags...)
    //handle configErr

    // load default widgets
    opentsgwidgets.AddBuiltinWidgets(otsg)
    // load our shiny nex example widget
    otsg.Handle(example.WidgetType, example.Schema, example.Config{})
    // run opentsg
    otsg.Run(*outputmnt)
```
