// Package tsg combines the core and widgets to draw the valeus for each frame
package tsg

import (
	"context"
	"fmt"
	"image/draw"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/mrmxf/opentsg-modules/opentsg-core/canvaswidget"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
)

type syncmap struct {
	syncer *sync.Mutex
	data   map[string]any
}

// OpenTSG is the engine for running openTSG
type OpenTSG struct {
	internal   context.Context
	framecount int

	// New Wave of handlers
	handlers    map[string]hand
	middlewares []func(Handler) Handler
	// 
	searchMiddleware []func(Search)Search
	encoders    map[string]Encoder
	// runner configuration
	runnerConf RunnerConfiguration
}

// RunnerConfiguration is the set up for the internal runners
// of openTSG
type RunnerConfiguration struct {
	// RunnerCount is the amount of runners (go routines)
	// that openTSG can use at anyone time
	RunnerCount int
	// Enable the profiler
	ProfilerEnabled bool
}

type hand struct {
	schema  []byte
	handler Handler
}

// BuildOpenTSG creates the OpenTSG engine.
// It is configured by an input json file and any profile set up information.
func BuildOpenTSG(inputFile string, profile string, debug bool, runnerConf *RunnerConfiguration, httpsKeys ...string) (*OpenTSG, error) {
	cont, framenumber, configErr := core.FileImport(inputFile, profile, debug, httpsKeys...)

	if configErr != nil {
		return nil, configErr
	}

	if runnerConf == nil {
		runnerConf = &RunnerConfiguration{RunnerCount: 1}
	}

	// stop negative runners appearing
	// and just locking everything up
	if runnerConf.RunnerCount < 1 {
		runnerConf.RunnerCount = 1
	}

	opentsg := &OpenTSG{internal: cont, framecount: framenumber,
		handlers:   map[string]hand{},
		encoders:   map[string]Encoder{},
		runnerConf: *runnerConf}

	// set up a canvaswidget handler, that runs empty
	opentsg.HandleFunc(canvaswidget.WType, HandlerFunc(func(_ Response, _ *Request) {}))

	return opentsg, nil
}

// NameSave is the extensions and encode function struct
// of a file type, to be used in tandem with the AddCustomSaves function.
type NameSave struct {
	Extension      string
	EncodeFunction func(io.Writer, draw.Image, int) error
}

// inTo4 converts an integer to 4 digit frame string number
func intToLength(num, length int) string {
	s := strconv.Itoa(num)
	buf0 := strings.Repeat("0", length-len(s))
	s = buf0 + s

	return s
}

// micro to Milli returns a format string of mictro sends as 00000.0 in milliseconds
func microToMili(duration int64) string {
	switch {
	case duration > 999999949: //99999949:
		return "99999.9" //   "99999.9"
	case duration < 50:
		return "000000.0" // "00000.0"
	default:
		// split the millisecond and micro second components
		base := time.Duration(duration).Truncate(time.Duration(time.Microsecond))
		decimal := duration - int64(base)
		dec := math.Round(float64(decimal) / 100)

		// check if it rounds up a whole number
		if dec == 10 {
			base += 1000
			dec = 0
		}
		bstring := intToLength(int(base)/1000, 6)

		return fmt.Sprintf("%s.%v", bstring, dec)
	}
}

// metaHook extracts all the user chosen metadata from a frame and its context.
func metaHookHandle(input draw.Image, c *context.Context) (map[string]any, error) {
	metaDataMap := make(map[string]any)

	if canvaswidget.GetMetaAverage(*c) {
		metaDataMap["Average Image Colour"] = averageCalc(input)
	}

	if canvaswidget.GetMetaConfiguration(*c) {
		metaDataMap["Frame Configuration"] = extractMetadata(c, "", "")
	}
	// return some hook stats

	return metaDataMap, nil
}

func averageCalc(targetImg draw.Image) any {
	count := 0
	b := targetImg.Bounds().Max
	R, G, B := 0, 0, 0
	for x := 0; x < b.X; x++ {
		for y := 0; y < b.Y; y++ {
			// these return the rgb16 value
			r, g, b, _ := targetImg.At(x, y).RGBA()
			R += int(r)
			G += int(g)
			B += int(b)
			count++
		}
	}

	type AverageImageColour struct {
		R int `yaml:"R"`
		G int `yaml:"G"`
		B int `yaml:"B"`
	}

	return AverageImageColour{R / count, G / count, B / count}

}

// Unmarshal unmarshals creates a function that unmarsahals yaml bytes
// into the handler type. This must be initialised with a struct.
func Unmarshal(han Handler) func(input []byte) (Handler, error) {

	return func(input []byte) (Handler, error) {
		// copy the underlying type to generate a new value
		// that points to the type that implements the handler method and not
		// just the handler method itself
		v := reflect.New(reflect.TypeOf(han))
		err := yaml.Unmarshal(input, v.Interface())

		if err != nil {
			return nil, err
		}

		return v.Interface().(Handler), nil
	}
}
