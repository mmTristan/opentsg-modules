package tsg

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/mrmxf/opentsg-modules/opentsg-core/canvaswidget"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/widgets"
	"github.com/mrmxf/opentsg-modules/opentsg-core/credentials"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"gopkg.in/yaml.v3"
)

// Handler is the OpenTSG interface for generating widgets
type Handler interface {
	Handle(Response, *Request)
}

// HandleFunc implements the Handler interface as a standalone functions
type HandlerFunc func(Response, *Request)

// Handle implements the handle method for functions
func (f HandlerFunc) Handle(resp Response, req *Request) {
	f(resp, req)
}

// HandleFunc registers the handler function for the given widget type in the
// openTSG engine
func (o OpenTSG) HandleFunc(wType string, handler HandlerFunc) {
	// set up router here

	o.Handle(wType, []byte("{}"), handler)

}

// Handle registers the handler class for the given widget type in the
// openTSG engine, with an accompanying json schema.
func (o OpenTSG) Handle(wType string, schema []byte, handler Handler) {

	if _, ok := o.handlers[wType]; ok {
		panic(fmt.Sprintf("The widget type %s has already been declared", wType))
	}

	// set a schema if one is given
	if schema == nil {
		schema = []byte("{}")
	}
	// do some checking for invalid characters, if there
	// are any

	o.handlers[wType] = hand{schema: schema, handler: handler}
}

// Use appends a middleware handler to the OpenTSG middleware stack.
func (o *OpenTSG) Use(middlewares ...func(Handler) Handler) {
	o.middlewares = append(o.middlewares, middlewares...)
}

// Request contains all the information sent to a widget handler
// for generating an image.
//
// It has methods for interacting with the core of openTSG.
// As well as set up for the image
type Request struct {
	// For http handlers etc
	RawWidgetYAML json.RawMessage
	JobID         string
	// the properties of the patch to be made
	PatchProperties PatchProperties
	FrameProperties FrameProperties

	// Helper functions that communicate with the engine
	// exported as methods
	// these are not exported as a json for http requests
	// the context is passed to widgets for this
	// offer a default for the text box search

	searchWithCredentials func(URI string) ([]byte, error)
	getWidgetMetadata     func(alias, dotpath string) any
}

// SearchWithCredentials searches a URI utilising any login credentials used
// when setting up openTSG.
//
// If the URI does not require any credentials then they are not used.
func (r Request) SearchWithCredentials(URI string) ([]byte, error) {
	if r.searchWithCredentials == nil {
		return credentials.GetWebBytes(nil, URI)
	}

	return r.searchWithCredentials(URI)
}

// GenerateSubImage generates an image of area bounds, that matches the type of image given to it.
//
// Note it will not work with custom draw.Image types
func (r Request) GenerateSubImage(baseImg draw.Image, bounds image.Rectangle) draw.Image {

	switch img := baseImg.(type) {
	case *colour.NRGBA64:
		return colour.NewNRGBA64(img.Space(), bounds)
	default:

		return image.NewNRGBA64(bounds)
	}
}

// GetWidget searches the metadata of a frame, for the alias and then recursivley searches through the keys to find nested values.
// The keys are a dotpath, if the key is invalid a nil value is returned.
// if the alias is "" then all the metadata is returned.
func (r Request) GetWidgetMetadata(alias, metadataField string) any {
	if r.getWidgetMetadata == nil {
		return nil
	}

	return r.getWidgetMetadata(alias, metadataField)
}

// PatchProperties contains the unique properties for
// the patch the widget is generating.
type PatchProperties struct {
	WidgetType string
	// the name of the widget call,
	// it is the full dot path
	WidgetFullID string
	// Dimensions and locations
	Dimensions  image.Rectangle
	TSGLocation image.Point
	// Geometry contains any tsig information
	Geometry    []gridgen.Segmenter
	ColourSpace colour.ColorSpace
}

// FrameProperties contains the overall properties for the
// frame that is currently being generated.
type FrameProperties struct {
	FrameNumber int
	// Where was openTSG called from
	// Often used for finding local files
	WorkingDir string
	//
	FrameDimensions image.Point
}

type Response interface {
	// Write a response to signal
	// the end of the widget and to handle any errors.
	// on success use the tsg.WidgetSuccess status code
	Write(status StatusCode, message string)

	// Return the base image to be handled.
	// Prevents overwriting the original draw.Image
	// with custom types
	BaseImage() draw.Image
}

// response implements the Response interface
type response struct {
	baseImg draw.Image
	status  StatusCode
	message string
}

// write to the response struct
func (r *response) Write(status StatusCode, message string) {
	// nothing is written at the moment
	r.status = status
	r.message = message
}

// return the base image to handle
func (r *response) BaseImage() draw.Image {
	return r.baseImg
}

// StatusCode is the OpenTSG status code
type StatusCode float64

const (
	FrameSuccess    = StatusCode(200.003)
	WidgetSuccess   = StatusCode(200.001)
	SaveSuccess     = StatusCode(200.002)
	WidgetNotFound  = StatusCode(404.001)
	EncoderNotFound = StatusCode(404.002)
	Profiler        = StatusCode(100.999)

	WidgetError   = StatusCode(500.001)
	WidgetWarning = StatusCode(400.001)
)

// ensure the final 3 digits are printed
func (s StatusCode) String() string {
	return fmt.Sprintf("%.3f", s)
}

// Legacy runs the old OpenTSG handler methods
// This version is currently ony used for testing.
type Legacy struct {
	// the location of the loader
	FileLocation string `json:"fileLocation" yaml:"fileLocation"`
	// mnt is the mount point of the folder
	MNT string `json:"mnt" yaml:"mnt"`
}

func (l Legacy) Handle(resp Response, req *Request) {

	otsg, err := BuildOpenTSG(l.FileLocation, "", true, nil)

	if err != nil {

		resp.Write(500, err.Error())
		return
	}

	// run the old program as normal
	otsg.Draw(true, l.MNT, "stdout")

	resp.Write(200, "")
}

// logErrors runs the internal errors as a handler
// used for dumping the error logs
func (tsg *OpenTSG) logErrors(code StatusCode, frameNumber int, jobId string, errors ...error) {
	errHan := HandlerFunc(func(resp Response, req *Request) {
		resp.Write(code, string(req.RawWidgetYAML))
	})
	errs := chain(tsg.middlewares, errHan)
	// call all errors so they are just logged
	for _, err := range errors {
		errs.Handle(&response{}, &Request{RawWidgetYAML: json.RawMessage(err.Error()),
			JobID:           jobId,
			PatchProperties: PatchProperties{WidgetFullID: "core.tsg"},
			FrameProperties: FrameProperties{FrameNumber: frameNumber},
		})
	}
}

// Run starts the OpenTSG engine, it runs every frame given
// to it from the set up file.
func (tsg *OpenTSG) Run(mnt string) {
	imageNo := tsg.framecount

	// wait for every frame to run before exiting the lopp
	var wg sync.WaitGroup
	wg.Add(tsg.framecount)

	// hookdata is a large map that contains all the metadata across the run.
	//var locker sync.Mutex
	//hookdata := syncmap{&locker, make(map[string]any)}

	// runFile := time.Now().Format("2006-01-02T15:04:05")

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
			jobID := gonanoid.MustID(16)
			monit := monitor{frameNo: frameNo, jobID: jobID}
			genMeasure := time.Now()
			saveTime := int64(0)
			// new log here for each frame

			// defer the progress bar message to use the values at the end of the "function"
			// the idea is for them to auto update
			defer func() {
				tsg.logErrors(FrameSuccess, frameNo, jobID,
					fmt.Errorf("generating frame %v/%v, gen: %v ms, save: %sms, errors:%v", frameNo, imageNo-1,
						microToMili(int64(time.Since(genMeasure).Microseconds())), microToMili(saveTime), monit.ErrorCount),
				)
				// add the log to the cache channel

			}()

			// update metadata to be included in the frame context
			frameConfigCont, errs := core.FrameWidgetsGeneratorHandle(tsg.internal, frameNo)

			// this is important for showing missed widget updates
			// log the errors
			if len(errs) > 0 {
				tsg.logErrors(404, frameNo, jobID, errs...)
				monit.incrementError(len(errs))
			}

			frameContext := &frameConfigCont
			errs = canvaswidget.LoopInitHandle(frameContext)

			if len(errs) > 0 {
				// log.Fatal
				tsg.logErrors(500, frameNo, jobID, errs...)
				monit.incrementError(len(errs))
				// frameWait.Done() //the frame weight is returned when the programs exit, or the frame has been generated
				return // continue // skip to the next frame number
			}
			// generate the canvas of type image.Image
			canvas, err := gridgen.GridGen(frameContext, core.GetDir(*frameContext),
				gridgen.FrameConfiguration{

					Rows:       canvaswidget.GetGridRows(*frameContext),
					Cols:       canvaswidget.GetGridColumns(*frameContext),
					LineWidth:  canvaswidget.GetLWidth(*frameContext),
					ImageSize:  canvaswidget.GetPictureSize(*frameContext),
					CanvasType: canvaswidget.GetCanvasType(*frameContext),
					CanvasFill: canvaswidget.GetFillColour(*frameContext),
					LineColour: canvaswidget.GetLineColour(*frameContext),
					ColorSpace: canvaswidget.GetBaseColourSpace(*frameContext),
					Geometry:   canvaswidget.GetGeometry(*frameContext),
					BaseImage:  canvaswidget.GetBaseImage(*frameContext),
				})
			if err != nil {
				tsg.logErrors(500, frameNo, jobID, err)
				monit.incrementError(1)

				return // continue // skip to the next frame number
			}

			// generate all the widgets
			tsg.widgetHandle(frameContext, canvas, &monit)

			// get the metadata and add it onto the map for this frame
			// @TODO update with the new metadata context
			/*md, _ := metaHook(canvas, frameContext)
			if len(md) != 0 { // only save if there actually is metadata
				i4 := intToLength(frameNo, 4)
				hookdata.syncer.Lock()
				hookdata.data[fmt.Sprintf("frame %s", i4)] = md
				hookdata.syncer.Unlock()
			}*/

			/*transformation station here where images can be moved to carved bits etc*/

			// save the image
			saveMeasure := time.Now()
			carves := gridgen.Carve(frameContext, canvas, canvaswidget.GetFileName(*frameContext))
			for _, carvers := range carves {
				// save.CanvasSave(canvas, canvaswidget.GetFileName(*frameContext), canvaswidget.GetFileDepth(*frameContext), mnt, i4, debug, frameLog)
				tsg.canvasSave2(carvers.Image, carvers.Location, canvaswidget.GetFileDepth(*frameContext), mnt, &monit)
			}
			saveTime = time.Since(saveMeasure).Microseconds()

		}()
		frameWait.Wait()

	}
	wg.Wait()
	fmt.Println("")

	/*

		move to a metadatahandler function
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

	*/
}

// CanvasSave saves the file according to the extensions provided
// the name add is for debug to allow to identify images
func (tsg *OpenTSG) canvasSave2(canvas draw.Image, filename []string, bitdeph int, mnt string, monit *monitor) {
	for _, name := range filename {
		truepath, err := filepath.Abs(filepath.Join(mnt, name))
		if err != nil {
			monit.incrementError(1)
			tsg.logErrors(700, monit.frameNo, monit.jobID, err)

			continue
		}
		err = tsg.encodeFrame(truepath, canvas, bitdeph)
		if err != nil {
			monit.incrementError(1)
			tsg.logErrors(700, monit.frameNo, monit.jobID, err)
		}
	}
}

type monitor struct {
	frameNo    int
	ErrorCount int
	jobID      string
	sync.Mutex
}

func (m *monitor) incrementError(count int) {
	m.Lock()
	m.ErrorCount += count
	m.Unlock()
}

type profile struct {
	SetUp     time.Duration `json:"SetUpTime(ns)"`
	Handler   time.Duration `json:"WidgetRunTime(ns)"`
	Queue     time.Duration `json:"QueueTime(ns)"`
	Composite time.Duration `json:"CompositeTime(ns)"`
	Wtype     string        `json:"type"`
	WID       string        `json:"widgetID"`
	ZPosition int           `json:"ZPosition"`
}

// // update widgetHandle to make the choices for me
func (tsg *OpenTSG) widgetHandle(c *context.Context, canvas draw.Image, monit *monitor) {

	// set up the core context functions
	allWidgets := widgets.ExtractAllWidgetsHandle(c)
	c = MetaDataInit(*c)
	// add the validator last
	lineErrs := core.GetJSONLines(*c)
	webSearch := func(URI string) ([]byte, error) {
		return credentials.GetWebBytes(c, URI)
	}

	// get the widgtes to be used
	// and intialiae the metadata
	allWidgetsArr := make([]core.AliasIdentityHandle, len(allWidgets))
	for alias, data := range allWidgets {

		allWidgetsArr[alias.ZPos] = alias
		put(c, alias.FullName, data)

	}

	extractFunc := func(alias, field string) any {
		return extractMetadata(c, alias, field)

	}

	// set up the properties for all requests
	fp := FrameProperties{WorkingDir: core.GetDir(*c), FrameNumber: monit.frameNo, FrameDimensions: canvas.Bounds().Max}

	zPos := 0
	// sync tools for running the widgets async
	runPool := Pool{AvailableMemeory: tsg.runnerConf.RunnerCount, drawers: &drawers{drawQueue: make(map[int]drawQueue), currentZ: &zPos}}
	// wg for each widget
	var wg sync.WaitGroup
	wg.Add(len(allWidgets))
	// ensure z order
	// prevent race conditions writing to the canvas
	// zpos := 0
	// zPos := &zpos
	// var zPosLock sync.Mutex
	//	var canvasLock sync.Mutex

	for i := 0; i < len(allWidgets); i++ {

		// get a runner to run the widget
		runner, available := runPool.GetRunner()
		for !available {

			time.Sleep(10 * time.Millisecond)
			runner, available = runPool.GetRunner()
		}
		p := profile{ZPosition: i}
		setUpStart := time.Now()
		// run the widget async
		go func() {

			position := i
			defer runPool.PutRunner(runner)
			defer wg.Done()

			widg := allWidgets[allWidgetsArr[i]]
			widgProps := allWidgetsArr[i]
			p.WID = widgProps.FullName
			p.Wtype = widgProps.WType

			handlers, handlerExists := tsg.handlers[allWidgetsArr[i].WType]
			// make a function so the handler is returned
			// @TODO skip the handler and come back to it later

			var Han Handler
			var resp response
			req := Request{JobID: gonanoid.MustID(16), getWidgetMetadata: extractFunc, PatchProperties: PatchProperties{WidgetFullID: widgProps.FullName, WidgetType: widgProps.WType}}
			var gridCanvas, mask draw.Image
			var imgLocation image.Point

			defer func() {

				// @TODO improve the handling of the object
				if tsg.runnerConf.ProfilerEnabled {

					if req.PatchProperties.WidgetType == "" {
						req.PatchProperties.WidgetType = widgProps.WType
						req.PatchProperties.WidgetFullID = widgProps.FullName
					}

					profiler := chain(tsg.middlewares, HandlerFunc(func(r1 Response, r2 *Request) {
						out, _ := json.Marshal(p)
						r1.Write(Profiler, string(out))
					}))
					profiler.Handle(&resp, &req)
				}
			}()

			// run a set up function that can return early
			// to make the handler just spit out the error
			func() {
				// ensure the chain is always kept
				defer func() {
					Han = chain(tsg.middlewares, Han)
					p.SetUp = time.Since(setUpStart)
				}()

				if !handlerExists {
					Han = GenErrorHandler(WidgetNotFound,
						fmt.Sprintf("No handler found for widgets of type \"%s\" for widget path \"%s\"", widgProps.WType, widgProps.FullName))
					return
				}

				var err error
				switch hdler := handlers.handler.(type) {
				// don't parse, as it will break
				// just run the function
				case HandlerFunc:
					Han = hdler
				default:
					Han, err = Unmarshal(handlers.handler)(widg)
				}

				if err != nil {
					Han = GenErrorHandler(400, err.Error())
					return
				}

				gridCanvas, imgLocation, mask, err = widgProps.Loc.GridSquareLocatorAndGenerator(c)

				// when the function am error is returned,
				// the function just becomes return an error
				if err != nil {
					Han = GenErrorHandler(400, err.Error())
					return
				}

				flats, err := widgProps.Loc.GetGridGeometry(c)
				if err != nil {
					Han = GenErrorHandler(400, err.Error())
					return
				}

				// do some colour stuff here

				// set up the requests
				// and chain the middleware for the handler
				pp := PatchProperties{WidgetType: widgProps.WType,
					WidgetFullID: widgProps.FullName,
					Dimensions:   gridCanvas.Bounds(),
					TSGLocation:  imgLocation, Geometry: flats,
					ColourSpace: widgProps.ColourSpace}
				//	Han, err := Unmarshal(handlers.handler)(widg)
				resp = response{baseImg: gridCanvas}
				req.FrameProperties = fp
				req.RawWidgetYAML = widg
				req.searchWithCredentials = webSearch
				req.PatchProperties = pp

				// chain that middleware at the last second?
				validatorMid := jSONValidator(lineErrs, handlers.schema, widgProps.FullName)
				Han = chain([]func(Handler) Handler{validatorMid}, Han)

			}()

			var canvasArea image.Rectangle
			if gridCanvas != nil {
				canvasArea = gridCanvas.Bounds().Add(imgLocation)
			}

			// log the area so other widgets can go while
			// the handler is running
			runPool.LogDrawArea(position, canvasArea)

			handleStart := time.Now()
			if widgProps.WType != "builtin.canvasoptions" {
				// RUN the widget
				Han.Handle(&resp, &req)
			}
			p.Handler = time.Since(handleStart)

			// wait until it is the widgets turn

			queue := time.Now()
			/*
				zPosLock.Lock()
				widgePos := *zPos
				zPosLock.Unlock()
				for widgePos != position {
					time.Sleep(time.Millisecond * 10)
					zPosLock.Lock()
					widgePos = *zPos
					zPosLock.Unlock()
				}
			*/
			// queue until the widget can run
			runPool.queue(runner, position, canvasArea)
			p.Queue = time.Since(queue)

			// only draw the image if
			// no errors occurred running the handler
			if resp.status == 200 || resp.status == WidgetSuccess {
				compostion := time.Now()
				//	canvasLock.Lock()
				colour.DrawMask(canvas, canvasArea, gridCanvas, image.Point{}, mask, image.Point{}, draw.Over)
				//	canvasLock.Unlock()
				p.Composite = time.Since(compostion)
				// else if there's been an error
			} else if widgProps.WType != canvaswidget.WType {
				// error of some sort from somewhere
				monit.incrementError(1)
			}

			// signal that the widget has finished
			runPool.CompleteZ(position)
			/*
				zPosLock.Lock()
				// update zpos regardless
				*zPos++
				zPosLock.Unlock()
			*/

		}()
	}

	wg.Wait()

}

type drawQueue struct {
	drawn bool
	area  image.Rectangle
}

type drawers struct {
	currentZ *int
	sync.Mutex
	drawQueue map[int]drawQueue
}

func GenErrorHandler(code StatusCode, errMessage string) Handler {
	return HandlerFunc(func(r Response, _ *Request) {
		r.Write(code, errMessage)
	})
}

func (p *Pool) CompleteZ(z int) {
	p.drawers.Lock()
	// @TODO check for doubles
	mid := p.drawers.drawQueue[z]
	mid.drawn = true
	p.drawers.drawQueue[z] = mid

	if z == *p.drawers.currentZ {
		// increment x amount of times until
		// the next undrawn widget is reached
		for p.drawers.drawQueue[*p.drawers.currentZ].drawn {
			*p.drawers.currentZ++
		}
	}
	p.drawers.Unlock()

}

func (p *Pool) LogDrawArea(position int, area image.Rectangle) {
	p.drawers.Lock()
	// @TODO check for doubles
	p.drawers.drawQueue[position] = drawQueue{area: area}
	p.drawers.Unlock()
}

func (p *Pool) queue(runner poolRunner, position int, area image.Rectangle) {

	// put the runner back with no memory
	// so the cache is still useable?
	clearPath := p.drawers.check(position, area)

	if !clearPath {
		// put the runner in the pool while queueing
		// no point putting it back if we are about to use it
		p.PutRunner(runner)
		// get it back out at the end
		defer func() {
			var available bool
			runner, available = p.GetRunner()
			for !available {

				time.Sleep(1 * time.Millisecond)
				runner, available = p.GetRunner()
			}
		}()
	}

	for !clearPath {
		time.Sleep(time.Millisecond * 1)
		clearPath = p.drawers.check(position, area)
	}

}

func (c *drawers) check(position int, area image.Rectangle) bool {
	c.Lock()
	widgePos := *c.currentZ
	c.Unlock()

	// do not bother checking areas underneath
	if widgePos == position {
		return true
	}

	clearPath := true
	// do not check against its own area
	for i := widgePos; i < position; i++ {
		c.Lock()
		under, ok := c.drawQueue[i]
		c.Unlock()
		if !ok {
			clearPath = false
			break
		}

		// already been drawn so area does
		// not matter
		if under.drawn {
			continue
		}

		// if any overlap then stop the search
		if area.Overlaps(under.area) {
			clearPath = false
			break
		}
	}
	return clearPath
}

type Pool struct {
	// keep at 1 at the moment
	AvailableMemeory int
	sync.Mutex
	drawers *drawers
}

func (p *Pool) GetRunner() (runner poolRunner, available bool) {

	p.Lock()
	defer p.Unlock()

	if p.AvailableMemeory > 0 {
		available = true
		runner = poolRunner{memory: 1}
		// remove the available runner
		p.AvailableMemeory--
	}

	return
}

func (p *Pool) PutRunner(run poolRunner) {

	p.Lock()
	defer p.Unlock()
	p.AvailableMemeory += run.memory

}

type poolRunner struct {
	memory int
}

// chain builds a http.Handler composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func chain(middlewares []func(Handler) Handler, endpoint Handler) Handler {

	// Return ahead of time if there aren't any middlewares for the chain
	if len(middlewares) == 0 {
		return endpoint
	}

	// Wrap the end handler with the middleware chain
	h := middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

type contextKey string

const (
	metaKey contextKey = "metadataKey"
)

///////////////////////
// Metadata Handling //
//////////////////////

// metadata is a map with a mutex to prevent concurrent read write to maps
type metadata struct {
	data map[string]map[string]interface{}
	mu   *sync.Mutex
}

// Put adds the information used to generate widget to the metadata
// is unexported to keep the information immutable
func put(c *context.Context, alias string, rawYaml []byte) error {

	// prevent concurrent writes
	imageGeneration := (*c).Value(metaKey).(metadata)
	imageGeneration.mu.Lock()
	defer imageGeneration.mu.Unlock()
	// map[string]map[string]map[string]interface{})

	var md map[string]any
	err := yaml.Unmarshal(rawYaml, &md)

	if err != nil {

		return fmt.Errorf("0201 Error inserting metadata %v", err)
	}
	imageGeneration.data[alias] = md

	// imageGeneration.data[widget] = alias

	return nil
}

// metaDataInit adds the metadata to the global context for that frame. This needs to be called
// before the widgets are run so that metadata can be stored.
func MetaDataInit(c context.Context) *context.Context {
	// MD is the metadata context of widget - alias - json(map[string] interface{})
	md := metadata{make(map[string]map[string]interface{}), &sync.Mutex{}}
	valC := context.WithValue(c, metaKey, md)
	return &valC
}

// Extract searches the metadata for an alias and then recursivley searches through the keys to find nested values.
// The keys are treated as a dotpath and will follow the path of any previous keys.
// if the alias is "" then all the metadata is returned.
func extractMetadata(c *context.Context, alias string, key string) interface{} {
	// order is widget type, alias map of json information
	imageGeneration := (*c).Value(metaKey).(metadata)
	// map[string]map[string]map[string]interface{})
	if alias == "" {

		return imageGeneration.data
	}
	start := imageGeneration.data[alias]

	keys := strings.Split(key, ".")
	if len(keys) != 0 {

		return mapToFace(start, keys...)
	}

	return start
}

func mapToFace(find map[string]any, keys ...string) interface{} {

	if find == nil {

		return find
	}

	if len(keys) == 0 {

		return find
	}
	if find[keys[0]] == nil {

		return find[keys[0]]
	}

	switch found := find[keys[0]].(type) {
	case map[string]interface{}:
		if len(keys) == 1 {

			return found
		}

		// loop through until all the keys are found or they aren't an interface
		return mapToFace(found, keys[1:]...)

	default:
		return found
	}

}
