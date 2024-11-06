package tsg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"log/slog"
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
	otsg, err := BuildOpenTSG("./testdata/testloader.json", "", true)

	first := func(next Handler) Handler {
		return HandlerFunc(func(r1 Response, r2 *Request) {
			fmt.Println("fitrst")

			next.Handle(r1, r2)
		})
	}

	mid := func(next Handler) Handler {
		return HandlerFunc(func(r1 Response, r2 *Request) {
			fmt.Println("SOME THING FROM CUSTOM MIDDLEWARE")

			next.Handle(r1, r2)
		})
	}
	otsg.Use(first, mid)
	otsg.Handle("builtin.legacy", []byte("{}"), Legacy{})
	otsg.HandleFunc("builtin.canvasoptions", func(r1 Response, r2 *Request) { fmt.Println("ring a ding") })
	fmt.Println(err)
	otsg.Run(true, "", "stdout")

}

/*
testMiddleware

run one that just ticks I am middleware that has run.
e.g. send to a log file

get one that proves the order runs in a certain way

*/

func TestHandlerAdditions(t *testing.T) {
	//	So(Ipanic, cv.ShouldPanic)

	otsg, err := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true)
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
}

func TestRaceConditions(t *testing.T) {
	otsg, _ := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true)
	otsg.Handle("test.fill", []byte("{}"), Filler{})

	fmt.Println(otsg.handler)
	otsg.Run(true, "", "stdout")
	// run with -race without the old code
	/*
		check the run
	*/
}

func TestMiddlewares(t *testing.T) {

	otsg, err := BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true)
	otsg.Handle("test.fill", []byte("{}"), Filler{})
	buf := bytes.NewBuffer([]byte{})
	jSlog := slog.NewJSONHandler(buf, &slog.HandlerOptions{})
	otsg.Use(Logger(slog.New(jSlog)))

	otsg.Run(true, "", "stdout")

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
				So(len(logEntries), ShouldEqual, 3)
				So(validLogs, ShouldBeTrue)
			})
		})
	})
	/*

		utilise the slogger to extract data points

	*/

	otsg, err = BuildOpenTSG("./testdata/handlerLoaders/loader.json", "", true)
	otsg.Handle("test.fill", []byte(`{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "type": "object",
    "properties": {
    },
    "required": [
        "fail"
    ]
}`), Filler{})
	buf = bytes.NewBuffer([]byte{})
	jSlog = slog.NewJSONHandler(buf, &slog.HandlerOptions{})
	otsg.Use(Logger(slog.New(jSlog)))

	otsg.Run(true, "", "stdout")

	// convert to count log entries
	logEntries = strings.Split(buf.String(), "\n")
	if logEntries[len(logEntries)-1] == "" {
		logEntries = logEntries[:len(logEntries)-1]
	}
	valid := messageValidator(logEntries, "400:0027 fail is required in unknown files please check your files for the fail property in the name cs.red\n")
	fmt.Println(logEntries)

	Convey("Checking the log handle runs the logs", t, func() {
		Convey("3 logs should be returned", func() {
			Convey("3 logs are returned denoting a successful run", func() {
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)

			})
		})
	})
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
	colour.Draw(r, r.Bounds(), &image.Uniform{fill}, image.Point{}, draw.Over)

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
