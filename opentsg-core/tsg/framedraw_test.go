package tsg

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colourgen"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLegacy(t *testing.T) {
	otsg, err := BuildOpenTSG("./testdata/testloader.json", "", true, nil)
	otsg.Handle("builtin.legacy", []byte("{}"), Legacy{})
	fmt.Println(err)
	otsg.Run("")

}

func TestTSIGWidget(t *testing.T) {
	// set up a fill handler that changes location each time

	area := `{
    "type": "test.fill",
    "grid": {
        "location": "%s"
    }
	}`

	areas := []string{"A1", "A0:a2", "r2c3", "R1C1:R3C3"}

	expectedArea := [][]gridgen.Segmenter{
		{{Name: "A001", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 10, Y: 10}}, Tags: []string{}, ImportPosition: 1}},
		{{Name: "A000", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 10, Y: 10}}, Tags: []string{}}, {Name: "A001", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 10}, Max: image.Point{X: 10, Y: 20}}, Tags: []string{}, ImportPosition: 1}},
		{},
		// some values are repeated across grids
		{{Name: "A000", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 10, Y: 10}}, Tags: []string{}, ImportPosition: 0},
			{Name: "A001", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 10}, Max: image.Point{X: 10, Y: 20}}, Tags: []string{}, ImportPosition: 1},
			{Name: "A002", Shape: image.Rectangle{Min: image.Point{X: 10, Y: 0}, Max: image.Point{X: 25, Y: 15}}, Tags: []string{}, ImportPosition: 2},
			{Name: "A003", Shape: image.Rectangle{Min: image.Point{X: 28, Y: 0}, Max: image.Point{X: 30, Y: 30}}, Tags: []string{}, ImportPosition: 3},
			{Name: "A004", Shape: image.Rectangle{Min: image.Point{X: 20, Y: 20}, Max: image.Point{X: 30, Y: 30}}, Tags: []string{}, ImportPosition: 4}}, {}}

	for i, a := range areas {
		f, fErr := os.Create("./testdata/tsigLoaders/tsigFill.json")
		_, wErr := f.Write([]byte(fmt.Sprintf(area, a)))

		otsg, err := BuildOpenTSG("./testdata/tsigLoaders/tsigLoader.json", "", true, nil)
		fmt.Println(err, fErr, wErr)
		var gotArea []gridgen.Segmenter
		otsg.HandleFunc("test.fill", HandlerFunc(func(r1 Response, r2 *Request) {

			gotArea = r2.PatchProperties.Geometry
		}))

		otsg.Run("")

		Convey("Calling openTSG with a tsig to ensure the correct values are returned", t, func() {
			Convey(fmt.Sprintf("Getting the tsigs in the grid area of \"%s\"", a), func() {
				Convey("no error is returned and we get the expected area", func() {
					So(fErr, ShouldBeNil)
					So(wErr, ShouldBeNil)
					So(err, ShouldBeNil)
					So(gotArea, ShouldResemble, expectedArea[i])

				})
			})
		})
	}

}

// TestHandlerAdditions checks the handler addition methods to
// catch the panics
func TestHandlerAdditions(t *testing.T) {
	//	So(Ipanic, cv.ShouldPanic)

	otsg, err := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, nil)
	otsg.Handle("test.fill", []byte("{}"), Filler{})

	Convey("Checking the handler panics when handles are redeclared", t, func() {
		Convey("adding test.fill as a function and an object", func() {
			Convey("both additions should panic", func() {
				So(err, ShouldBeNil)
				So(func() { otsg.Handle("test.fill", []byte("{}"), Filler{}) }, ShouldPanic)
				So(func() { otsg.HandleFunc("test.fill", Filler{}.Handle) }, ShouldPanic)
			})
		})
	})

	otsgEncoder, err := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, nil)
	AddBaseEncoders(otsgEncoder)

	Convey("Checking the tsg encoder handler panics when encoders are redeclared", t, func() {
		Convey("duplicating the encoders, with AddBaseEncoders", func() {
			Convey("the additions should panic", func() {
				So(err, ShouldBeNil)
				So(func() { AddBaseEncoders(otsgEncoder) }, ShouldPanic)
			})
		})
	})
}

func TestMethodFunctions(t *testing.T) {
	// @TODO test the response and request methods

	otsg, err := BuildOpenTSG("./testdata/testloader.json", "", true, nil)
	otsg.HandleFunc("builtin.legacy", HandlerFunc(func(r1 Response, r2 *Request) {

		r2.searchWithCredentials("")
	}))
	fmt.Println(err, "this err")
	otsg.Run("")
}

// Run with the -race flag to ensure no shenanigans occur
func TestRaceConditions(t *testing.T) {
	// run with plenty of runners to ensure all the go routines are running at once
	otsg, buildErr := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, &RunnerConfiguration{RunnerCount: 5})
	otsg.Handle("test.fill", []byte("{}"), Filler{})
	jSlog := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	AddBaseEncoders(otsg)
	otsg.Use(Logger(slog.New(jSlog)))
	otsg.Run("")

	genFile, _ := os.Open("./testdata/handlerLoaders/racer.png")
	// Decode to get the colour values
	baseVals, _ := png.Decode(genFile)
	// Assign the colour to the correct type of image NGRBA64 and replace the colour values
	genImage := image.NewNRGBA64(baseVals.Bounds())
	colour.Draw(genImage, genImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

	// Open the image to compare to
	controlFile, _ := os.Open("./testdata/handlerLoaders/expectedRace.png")
	// Decode to get the colour values
	controlVals, _ := png.Decode(controlFile)

	// Assign the colour to the correct type of image NGRBA64 and replace the colour values
	controlImage := image.NewNRGBA64(controlVals.Bounds())
	colour.Draw(controlImage, controlImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

	// Make a hash of the pixels of each image
	hnormal := sha256.New()
	htest := sha256.New()
	hnormal.Write(controlImage.Pix)
	htest.Write(genImage.Pix)

	Convey("Checking for race conditions", t, func() {
		Convey("running boxes on top of each other, that should alway layer red, green then blue", func() {
			Convey("No races occur and the picture matches the expected", func() {
				So(buildErr, ShouldBeNil)
				So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
			})
		})
	})
}

// JSONLog is the key fields of the json slogger
// for testing against.
type JSONLog struct {
	StatusCode string `json:"StatusCode"`
	WidgetID   string `json:"WidgetID"`
}

func TestMiddlewares(t *testing.T) {

	otsg, err := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, nil)
	otsg.Handle("test.fill", []byte("{}"), Filler{})
	AddBaseEncoders(otsg)
	buf := bytes.NewBuffer([]byte{})
	jSlog := slog.NewJSONHandler(buf, &slog.HandlerOptions{})
	otsg.Use(Logger(slog.New(jSlog)))
	otsg.Run("")

	var outMessage JSONLog
	jErr := json.Unmarshal(buf.Bytes(), &outMessage)

	// @TODO check the messages are correct
	Convey("Checking the log handler actually logs", t, func() {
		Convey("making a run at the log level info", func() {
			Convey("one log is returned denoting a successful run", func() {
				So(err, ShouldBeNil)
				So(jErr, ShouldBeNil)
				So(outMessage, ShouldResemble, JSONLog{WidgetID: "core.tsg", StatusCode: FrameSuccess.String()})
			})
		})
	})

	// set up the order
	otsg, err = BuildOpenTSG("./testdata/handlerLoaders/singleloader.json", "", true, nil)
	AddBaseEncoders(otsg)
	otsg.Handle("test.fill", []byte(`{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "type": "object",
    "properties": {
    },
    "required": [
        "fail"
    ]
}`), Filler{})

	first := func(next Handler) Handler {
		return HandlerFunc(func(r1 Response, r2 *Request) {
			r1.Write(500, "first")
			next.Handle(r1, r2)
		})
	}

	second := func(next Handler) Handler {
		return HandlerFunc(func(r1 Response, r2 *Request) {
			r1.Write(500, "second")
			next.Handle(r1, r2)
		})
	}
	orderLog := &testSlog{logs: make([]string, 0), level: slog.LevelError}
	otsg.Use(Logger(slog.New(orderLog)), first, second)
	otsg.Run("")

	Convey("Checking the middleware runs in the oder it is called", t, func() {
		Convey("the return of the logs are 3 messages in the order of, first, second and validator", func() {
			Convey("the logs match that order", func() {
				So(err, ShouldBeNil)
				So(orderLog.logs, ShouldResemble, []string{"first", "second",
					"0027 fail is required in unknown files please check your files for the fail property in the name blue,",
					"first", "second",
				})
			})
		})
	})

	otsg, err = BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, nil)
	AddBaseEncoders(otsg)
	otsg.Handle("test.fill", []byte(`{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "type": "object",
    "properties": {
    },
    "required": [
        "fail"
    ]
}`), Filler{})

	logArr := testSlog{logs: make([]string, 0), level: slog.LevelError}
	otsg.Use(Logger(slog.New(&logArr)))

	otsg.Run("")

	Convey("Checking the validator middleware returns errors", t, func() {
		Convey("running the json with a schema that will fail", func() {
			Convey("3 logs are returned denoting, each denoting a schema failure", func() {
				So(err, ShouldBeNil)
				So(logArr.logs, ShouldResemble, []string{"0027 fail is required in unknown files please check your files for the fail property in the name cs.blue,",
					"0027 fail is required in unknown files please check your files for the fail property in the name cs.green,",
					"0027 fail is required in unknown files please check your files for the fail property in the name cs.red,"})

			})
		})
	})
}

func TestMetadata(t *testing.T) {

	otsg, _ := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, nil)
	otsg.Handle("test.fill", []byte("{}"), Filler{})
	AddBaseEncoders(otsg)

	search := []string{"cs.blue", "cs.red", "cs.green", "cs.green"}
	fields := []string{"type", "grid", "mdObject", "grid.location"}
	expected := []any{"test.fill", map[string]any{"location": "a0:d3"}, 400.003, "a0:e4"}

	for i, salias := range search {

		var result any

		extractor := func(next Handler) Handler {
			return HandlerFunc(func(r Response, req *Request) {

				// for this middleware only run on the searched widget
				// to ignore other results which we may not want
				if req.PatchProperties.WidgetFullID == salias {
					result = req.GetWidgetMetadata(salias, fields[i])

				}
				next.Handle(r, req)
			})
		}

		otsg.Use(extractor)
		otsg.Run("")

		// @TODO check the messages are correct
		Convey("Checking the metadata extraction function runs", t, func() {
			Convey(fmt.Sprintf("Searching the metadata for the field of \"%s\"", fields[i]), func() {
				Convey("The expected value is returned", func() {
					So(result, ShouldResemble, expected[i])
				})
			})
		})
	}
}

type Filler struct {
	Fill string `json:"fill" yaml:"fill"`
}

func (f Filler) Handle(r Response, _ *Request) {

	fill := colourgen.HexToColour(f.Fill, colour.ColorSpace{})
	colour.Draw(r.BaseImage(), r.BaseImage().Bounds(), &image.Uniform{fill}, image.Point{}, draw.Over)

	r.Write(200, "success")
}

func TestMarshallHandler(t *testing.T) {

	testHandlers := map[string]Handler{
		"dummyHandler":       dummyHandler{},
		"secondDummyHandler": secondDummyHandler{},
	}

	expected := map[string]Handler{
		"dummyHandler":       &dummyHandler{"testInput"},
		"secondDummyHandler": &secondDummyHandler{"testInput"},
	}

	tragets := []string{"dummyHandler", "secondDummyHandler"}

	for _, target := range tragets {

		gotHandler, err := Unmarshal(testHandlers[target])([]byte(`{"input":"testInput"}`))

		Convey("Checking the unmarshaling of bytes to method structs", t, func() {
			Convey(fmt.Sprintf("Unmarshaling bytes to a struct of %v ", reflect.TypeOf(testHandlers[target])), func() {
				Convey("No error is returned and the struct is populated as expected", func() {
					So(err, ShouldBeNil)
					So(gotHandler, ShouldResemble, expected[target])
				})
			})
		})

	}
}

func TestErrors(t *testing.T) {

	errors := []string{
		`{
    "type": "test.fills",
    "grid": {
        "location": "a0:f5"
    },
    "fill":"#0000ff"
}`,
		`{
    "type": "test.fill",
    "grid": {
        "location": "a"
    },
    "fill":"#0000ff"
}`}

	expectedErrs := []string{"No handler found for widgets of type \"test.fills\" for widget path \"err\"",
		"0046 a is not a valid grid alias"}

	for i, e := range errors {
		f, fErr := os.Create("./testdata/handlerLoaders/err.json")
		_, wErr := f.Write([]byte(e))

		otsg, err := BuildOpenTSG("./testdata/handlerLoaders/errLoader.json", "", true, nil)
		otsg.Handle("test.fill", []byte(`{}`), Filler{})

		orderLog := &testSlog{logs: make([]string, 0), level: slog.LevelWarn}
		otsg.Use(Logger(slog.New(orderLog)))
		AddBaseEncoders(otsg)
		otsg.Run("")

		Convey("Calling openTSG with a widget that deliberately fails", t, func() {
			Convey(fmt.Sprintf("using a json of \"%s\"", e), func() {
				Convey(fmt.Sprintf("An error of message \"%s\" is returned", expectedErrs[i]), func() {
					So(fErr, ShouldBeNil)
					So(wErr, ShouldBeNil)
					So(err, ShouldBeNil)
					So(orderLog.logs, ShouldResemble, []string{expectedErrs[i]})
				})
			})
		})
	}
}

func TestQueue(t *testing.T) {

	areas := []image.Rectangle{
		image.Rect(0, 0, 10, 10),
		image.Rect(50, 50, 60, 60),
	}
	firstRun := []bool{
		true, false,
	}
	message := []string{
		"running a check, before the first widget entry has been recorded. Pausing all future widgets",
		"running a widget that overlaps an area, that has not been drawn to",
	}

	for i, ba := range areas {
		out := setUpPoolRunner(0, firstRun[i], false)

		pass := out.drawers.check(3, ba)
		fmt.Println(out.drawers.drawQueue)
		Convey("Checking the z order handler stops premature widgets", t, func() {
			Convey(message[i], func() {
				Convey("A false value is returned stating it is not the widgets turn", func() {
					So(pass, ShouldBeFalse)
				})
			})
		})
	}

	runStatus := []bool{
		true, false,
	}

	message = []string{
		"running the next when all previous widgets have been written to.",
		"running a widget that has no overlap",
	}

	for i, ba := range areas {
		out := setUpPoolRunner(0, true, runStatus[i])

		pass := out.drawers.check(3, ba)
		fmt.Println(out.drawers.drawQueue)
		Convey("Checking the z order handler allows widgets that do not clash", t, func() {
			Convey(message[i], func() {
				Convey("A true value is returned stating it is the widgets turn", func() {
					So(pass, ShouldBeTrue)
				})
			})
		})
	}

}

func setUpPoolRunner(zPos int, firstWidgetPresent, runStatus bool) *Pool {

	layers := map[int]drawQueue{

		1: {drawn: runStatus, area: image.Rect(10, 10, 20, 20)},
		2: {drawn: runStatus, area: image.Rect(20, 20, 30, 30)},
	}

	if firstWidgetPresent {
		layers[0] = drawQueue{drawn: runStatus, area: image.Rect(0, 0, 10, 10)}
	}

	return &Pool{drawers: &drawers{currentZ: &zPos, drawQueue: layers}}
}

type dummyHandler struct {
	Input string `json:"input"`
}

func (d dummyHandler) Handle(resp Response, req *Request) {
}

type secondDummyHandler struct {
	Input string `json:"input"`
}

func (d secondDummyHandler) Handle(resp Response, req *Request) {
}

// test log is a struct for piping logs into
// an array.
// not thread safe and just something dumb for tests
type testSlog struct {
	logs  []string
	level slog.Level
}

func (ts *testSlog) Enabled(context.Context, slog.Level) bool {
	return true
}

func (ts *testSlog) Handle(_ context.Context, rec slog.Record) error {

	if rec.Level >= ts.level {
		ts.logs = append(ts.logs, rec.Message)
	}

	return nil
}

func (ts *testSlog) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ts
}

func (ts *testSlog) WithGroup(name string) slog.Handler {
	return ts
}
