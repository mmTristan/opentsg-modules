// Package tsg combines the core and widgets to draw the valeus for each frame
package tsg

import (
	"context"
	"fmt"
	"image/draw"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/mrmxf/opentsg-modules/opentsg-core/canvaswidget"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
)

type syncmap struct {
	syncer *sync.Mutex
	data   map[string]any
}

// OpenTSG is the engine for running openTSG
type OpenTSG struct {
	internal   context.Context
	framecount int
	// to be trimmed
	customWidgets []func(chan draw.Image, bool, *context.Context, *sync.WaitGroup, *sync.WaitGroup, *errhandle.Logger)
	customSaves   map[string]func(io.Writer, draw.Image, int) error

	// New Wave of handlers
	handlers    map[string]hand
	middlewares []func(Handler) Handler
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
		customWidgets: baseWidgets(),
		handlers:      map[string]hand{},
		encoders:      map[string]Encoder{},
		customSaves:   baseSaves(),
		runnerConf:    *runnerConf}

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

/*
AddCustomSaves allows for custom save functions to be added to the opentsg object.

e.g. a save function may look like this

	func WriteJPEGFile(file *os.File, img draw.Image, empty int) error {

		return jpeg.Encode(file, img, &jpeg.Options{Quality: 100})
	}

and then would be  to the tsg object like this

	tpg.AddCustomSaves([]tsg.NameSave{{Extension: "JPEG", SaveFunction: WriteJPEGFile}
*/
func (tsg *OpenTSG) AddCustomSaves(customSaves []NameSave) {
	// need name and save function
	// TODO:emit warnings

	for _, save := range customSaves {
		tsg.customSaves[strings.ToUpper(save.Extension)] = save.EncodeFunction
	}

}

// Add CustomWidgets allows for custom widget functions to be run from opentsg. Without going into the internals of the opentsg and changing things up.
// To understand the design of the widgets function, please check the layout of the widget module.
func (tsg *OpenTSG) AddCustomWidgets(widgets ...func(chan draw.Image, bool, *context.Context, *sync.WaitGroup, *sync.WaitGroup, *errhandle.Logger)) {
	tsg.customWidgets = append(tsg.customWidgets, widgets...)
}

// Draw generates the images for each array section of the json array and applies it to the test card grid.
// This is the legacy version of OTSG
func (tsg *OpenTSG) Draw(debug bool, mnt, logType string) {
	imageNo := tsg.framecount

	// wait for every frame to run before exiting the lopp
	var wg sync.WaitGroup
	wg.Add(tsg.framecount)

	logs := make(chan *errhandle.Logger, imageNo)

	// hookdata is a large map that contains all the metadata across the run.
	var locker sync.Mutex
	hookdata := syncmap{&locker, make(map[string]any)}

	runFile := time.Now().Format("2006-01-02T15:04:05")

	for frameLoopNo := 0; frameLoopNo < imageNo; frameLoopNo++ {
		// make an internal function
		// so that a defer print statement can be used at the end of each frame generation
		// and for running as a go this reduces time by about 40%?
		frameNo := frameLoopNo
		var frameWait sync.WaitGroup
		frameWait.Add(1)
		go func() {
			defer wg.Done()
			defer frameWait.Done()

			genMeasure := time.Now()
			saveTime := int64(0)
			// new log here for each frame
			frameLog := errhandle.LogInit(logType, mnt)
			// defer the progress bar message to use the values at the end of the "function"
			// the idea is for them to auto update
			defer func() {
				fmt.Printf("\rGenerating frame %v/%v, gen: %v ms, save: %sms, errors:%v\n", frameNo, imageNo-1, microToMili(int64(time.Since(genMeasure).Microseconds())), microToMili(saveTime), frameLog.ErrorCount())
				// add the log to the cache channel
				logs <- frameLog
			}()

			// change the log prefix for each image we generate, make a logger for each one for concurrency at a later date
			i4 := intToLength(frameNo, 4)
			frameLog.SetPrefix(fmt.Sprintf("%v_", i4)) // update prefix to just be frame number
			// update metadata to be included in the frame context
			frameConfigCont, errs := core.FrameWidgetsGenerator(tsg.internal, frameNo)

			// this is important for showing missed widget updates
			for _, e := range errs {
				frameLog.PrintErrorMessage("W_CORE_opentsg_", e, true)
			}

			frameContext := widgethandler.MetaDataInit(frameConfigCont)
			errs = canvaswidget.LoopInit(frameContext)

			if len(errs) > 0 {

				// print all the errors
				for _, e := range errs {
					frameLog.PrintErrorMessage("F_CORE_opentsg_", e, debug)
				}
				// frameWait.Done() //the frame weight is returned when the programs exit, or the frame has been generated

				return // continue // skip to the next frame number
			}
			// generate the canvas of type image.Image
			canvas, err := gridgen.GridGen(frameContext, core.GetDir(*frameContext),
				gridgen.FrameConfiguration{

					Rows:       canvaswidget.GetGridRows(*frameContext),
					Cols:       canvaswidget.GetGridColumns(*frameContext),
					LineWidth:  canvaswidget.GetLWidth(*frameContext),
					FrameSize:  canvaswidget.GetPictureSize(*frameContext),
					CanvasType: canvaswidget.GetCanvasType(*frameContext),
					CanvasFill: canvaswidget.GetFillColour(*frameContext),
					LineColour: canvaswidget.GetLineColour(*frameContext),
					ColorSpace: canvaswidget.GetBaseColourSpace(*frameContext),
					Geometry:   canvaswidget.GetGeometry(*frameContext),
					BaseImage:  canvaswidget.GetBaseImage(*frameContext),
				})
			if err != nil {
				frameLog.PrintErrorMessage("F_CORE_opentsg_", err, debug)
				// frameWait.Done()
				return // continue // skip to the next frame number
			}

			// generate all the widgets
			tsg.widgetGen(frameContext, debug, canvas, frameLog)
			// frameWait.Done()

			// get the metadata and add it onto the map for this frame
			md, _ := metaHook(canvas, frameContext)
			if len(md) != 0 { // only save if there actually is metadata
				hookdata.syncer.Lock()
				hookdata.data[fmt.Sprintf("frame %s", i4)] = md
				hookdata.syncer.Unlock()
			}

			/*transformation station here where images can be moved to carved bits etc*/

			// save the image
			saveMeasure := time.Now()
			carves := gridgen.Carve(frameContext, canvas, canvaswidget.GetFileName(*frameContext))
			for _, carvers := range carves {
				// save.CanvasSave(canvas, canvaswidget.GetFileName(*frameContext), canvaswidget.GetFileDepth(*frameContext), mnt, i4, debug, frameLog)
				tsg.canvasSave(carvers.Image, carvers.Location, canvaswidget.GetFileDepth(*frameContext), mnt, i4, debug, frameLog)
			}
			saveTime = time.Since(saveMeasure).Microseconds()

		}()
		frameWait.Wait()

	}
	wg.Wait()
	fmt.Println("")

	if debug {
		// generate the metadata folder, if it has had any generated data
		if len(hookdata.data) != 0 {
			// write a better name for identfying
			metaLocation, _ := filepath.Abs(mnt + "./" + runFile + ".yaml")
			md, _ := os.Create(metaLocation)
			b, _ := yaml.Marshal(hookdata.data)
			md.Write(b)
		}
	}

	// flush the logs in the order they were cached in the channel
	// logs are flushed in batches of their frames
	for len(logs) > 0 {
		l := <-logs
		l.LogFlush()
	}
}

func baseWidgets() []func(chan draw.Image, bool, *context.Context, *sync.WaitGroup, *sync.WaitGroup, *errhandle.Logger) {
	return []func(chan draw.Image, bool, *context.Context, *sync.WaitGroup, *sync.WaitGroup, *errhandle.Logger){}
}

// each image is added to the base image
func (tsg *OpenTSG) widgetGen(c *context.Context, debug bool, canvas draw.Image, logs *errhandle.Logger) {

	// gridgen.ParamToCanvas can be changed depending on the coordinate system
	canvasChan := make(chan draw.Image, 1)
	// put the canvas in a channel to prevent race conditions as a pointer
	// it should be called to only be added then returned to the
	canvasChan <- canvas

	var wg sync.WaitGroup  // widget waitgroup
	var wgc sync.WaitGroup // context waitgroup

	// add new widgets to the list of widgets, new widgets can be plugged in
	// and the list can be amended.
	widgets := tsg.customWidgets
	wg.Add(len(widgets))
	wgc.Add(len(widgets))
	// loop through and run the widgets to glean the metadata and set their context for the run
	for _, w := range widgets {
		go w(canvasChan, debug, c, &wg, &wgc, logs)
	}

	// wait for all the widgets to have run before returning
	wg.Wait()
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
func metaHook(input draw.Image, c *context.Context) (map[string]any, error) {
	metaDataMap := make(map[string]any)

	// assign all the generated metadata here straight onto the map
	// wrap that as a function https://github.com/corona10/goimagehash
	/* TODO finish adding the hash
		if canvaswidget.GetMetaPhash(*c) {
			//make this a function and choose a phash to make
		//	fmt.Println(phash.DTC(input))
			g, _ := goimagehash.PerceptionHash(input)
		//	fmt.Println(g.GetHash())
			ph := imghash.NewPHash()
			bin := ph.Calculate(input)
			little := binary.LittleEndian.Uint64(bin)

	//		fmt.Println(little)
			big := binary.BigEndian.Uint64(bin)
	//		fmt.Println(big, "big")
	//		fmt.Println(imgsim.AverageHash(input).String())
		}*/

	if canvaswidget.GetMetaAverage(*c) {
		metaDataMap["Average Image Colour"] = averageCalc(input)
	}

	if canvaswidget.GetMetaConfiguration(*c) {
		metaDataMap["Frame Configuration"] = widgethandler.Extract(c, "")
	}
	// return some hook stats

	return metaDataMap, nil
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
func Unmarshal(Han Handler) func(input []byte) (Handler, error) {

	return func(input []byte) (Handler, error) {
		// copy the underlying type to generate a new value
		// that points to the type that implements the handler method and not
		// just the handler method itself
		v := reflect.New(reflect.TypeOf(Han))
		err := yaml.Unmarshal(input, v.Interface())

		if err != nil {
			return nil, err
		}

		return v.Interface().(Handler), nil
	}
}
