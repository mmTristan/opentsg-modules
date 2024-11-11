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
	"strings"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colourgen"
	. "github.com/smartystreets/goconvey/convey"
)

/*
func TestXxx(t *testing.T) {

		otsg := openTSG{widgets: map[string]func([]byte) (Generator, error){}} // map[string]func([]byte) (Generator, error){}}
		otsg.AddWidget("test", []byte("{}"), testWidget{})
		otsg.AddWidget("test2", []byte("{}"), testWidget2{})

		bs := []byte("{\"input\":\"yes\"}")
		//var x any
		x, err := otsg.widgets["test"](bs)

		fmt.Println(err)
		x.Generate(nil)

		x2, _ := otsg.widgets["test2"](bs)
		x2.Generate(nil)
	}
*/

func TestLegacy(t *testing.T) {
	otsg, err := BuildOpenTSG("./testdata/testloader.json", "", true, nil)
	otsg.Handle("builtin.legacy", []byte("{}"), Legacy{})
	otsg.HandleFunc("builtin.canvasoptions", func(r1 Response, r2 *Request) { fmt.Println("ring a ding") })
	fmt.Println(err)
	otsg.Run("")

}

/*
testMiddleware

run one that just ticks I am middleware that has run.
e.g. send to a log file

get one that proves the order runs in a certain way

*/

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

}

func TestRaceConditions(t *testing.T) {
	otsg, buildErr := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, &RunnerConfiguration{5})
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

func TestMiddlewares(t *testing.T) {

	otsg, err := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, nil)
	otsg.Handle("test.fill", []byte("{}"), Filler{})
	AddBaseEncoders(otsg)
	buf := bytes.NewBuffer([]byte{})
	jSlog := slog.NewJSONHandler(buf, &slog.HandlerOptions{})
	otsg.Use(Logger(slog.New(jSlog)))

	otsg.Run("")

	// convert to count log entries
	logEntries := strings.Split(buf.String(), "\n")
	if logEntries[len(logEntries)-1] == "" {
		logEntries = logEntries[:len(logEntries)-1]
	}

	validLogs := messageValidator(logEntries, "200:success")

	// @TODO check the messages are correct
	Convey("Checking the log handle runs the logs", t, func() {
		Convey("3 logs should be returned", func() {
			Convey("3 logs are returned denoting a successful run", func() {
				So(err, ShouldBeNil)
				So(len(logEntries), ShouldEqual, 4)
				So(validLogs, ShouldResemble, []string{})
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
			r1.Write(0, "first")
			next.Handle(r1, r2)
		})
	}

	second := func(next Handler) Handler {
		return HandlerFunc(func(r1 Response, r2 *Request) {
			r1.Write(0, "second")
			next.Handle(r1, r2)
		})
	}
	orderLog := &testSlog{logs: make([]string, 0)}
	otsg.Use(Logger(slog.New(orderLog)), first, second)
	otsg.Run("")

	Convey("Checking the middleware runs in the oder it is called", t, func() {
		Convey("the return of the logs are 3 messages in the order of, first, second and validator", func() {
			Convey("the logs match that order", func() {
				So(err, ShouldBeNil)
				So(orderLog.logs, ShouldResemble, []string{"0:first", "0:second",
					"400:0027 fail is required in unknown files please check your files for the fail property in the name blue,",
					"0:first", "0:second", "200:generating frame 0/0, gen: 000525.0 ms, save: 000408.8ms, errors:1",
				})
			})
		})
	})

	/*

	   test run order

	*/

	/*

		utilise the slogger to extract data points

	*/

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

	logArr := testSlog{logs: make([]string, 0)}
	otsg.Use(Logger(slog.New(&logArr)))

	otsg.Run("")

	// convert to count log entries
	logEntries = strings.Split(buf.String(), "\n")
	if logEntries[len(logEntries)-1] == "" {
		logEntries = logEntries[:len(logEntries)-1]
	}
	//valid := messageValidator(logEntries, "400:0027 fail is required in unknown files please check your files for the fail property in the name cs.red\n")
	//fmt.Println(logEntries)

	Convey("Checking the log handle runs the logs", t, func() {
		Convey("3 logs should be returned", func() {
			Convey("3 logs are returned denoting a successful run", func() {
				So(err, ShouldBeNil)
				So(logArr.logs, ShouldResemble, []string{"400:0027 fail is required in unknown files please check your files for the fail property in the name cs.blue,",
					"400:0027 fail is required in unknown files please check your files for the fail property in the name cs.green,",
					"400:0027 fail is required in unknown files please check your files for the fail property in the name cs.red,",
					"200:generating frame 0/0, gen: 000619.6 ms, save: 000475.6ms, errors:3"})

			})
		})
	})
}

func TestMetadata(t *testing.T) {

	otsg, _ := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true, nil)
	otsg.Handle("test.fill", []byte("{}"), Filler{})
	AddBaseEncoders(otsg)

	search := []string{"cs.blue", "cs.red", "cs.green"}
	fields := []string{"type", "grid", "mdObject"}
	expected := []any{"test.fill", map[string]any{"location": "a0:d3"}, 400.003}

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
		Convey("Checking the log handle runs the logs", t, func() {
			Convey("3 logs should be returned", func() {
				Convey("3 logs are returned denoting a successful run", func() {
					So(result, ShouldResemble, expected[i])
				})
			})
		})
	}
}

func messageValidator(messages []string, expectedMess string) any {

	type record struct {
		Message string `json:"msg"`
	}

	for _, mess := range messages {
		var r record
		json.Unmarshal([]byte(mess), &r)
		if r.Message != expectedMess {
			return fmt.Sprintf("Got message %s expected %s", r.Message, expectedMess)
		}
	}
	return true
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

	// have a set of jsons inserted into a runner
	/*
		these consist of bad json
		bad coordinates
		invalid gridgen for tsg? - maybe laters

	*/

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
}`,
	}

	expectedErrs := []string{"400:No handler found for widgets of type \"test.fills\" for widget path \"err\"",
		"400:0046 a is not a valid grid alias"}

	for i, e := range errors {
		f, fErr := os.Create("./testdata/handlerLoaders/err.json")
		_, wErr := f.Write([]byte(e))

		otsg, err := BuildOpenTSG("./testdata/handlerLoaders/errLoader.json", "", true, nil)
		otsg.Handle("test.fill", []byte(`{}`), Filler{})

		orderLog := &testSlog{logs: make([]string, 0)}
		otsg.Use(Logger(slog.New(orderLog)))
		AddBaseEncoders(otsg)
		otsg.Run("")

		Convey("Calling openTSG with a widget that deliberately fails", t, func() {
			Convey(fmt.Sprintf("using a json of \"%s\"", e), func() {
				Convey(fmt.Sprintf("An error of \"%s\" is returned", expectedErrs[i]), func() {
					So(fErr, ShouldBeNil)
					So(wErr, ShouldBeNil)
					So(err, ShouldBeNil)
					So(orderLog.logs, ShouldResemble, []string{expectedErrs[i], "200:generating frame 0/0, gen: 000463.4 ms, save: 000378.1ms, errors:1"})
				})
			})
		})
	}

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
	logs []string
}

func (ts *testSlog) Enabled(context.Context, slog.Level) bool {
	return true
}

func (ts *testSlog) Handle(_ context.Context, rec slog.Record) error {

	ts.logs = append(ts.logs, rec.Message)

	return nil
}

func (ts *testSlog) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ts
}

func (ts *testSlog) WithGroup(name string) slog.Handler {
	return ts
}
