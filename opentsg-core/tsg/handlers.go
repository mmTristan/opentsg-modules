package tsg

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
	"sync"
	"time"

	"github.com/mrmxf/opentsg-modules/opentsg-core/canvaswidget"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/widgets"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
)

type Handler interface {
	Handle(Response, *Request)
}

type HandlerFunc func(Response, *Request)

func (f HandlerFunc) Handle(resp Response, req *Request) {
	f(resp, req)
}

// HandleFunc registers the handler function for the given pattern in [DefaultServeMux].
// The documentation for [ServeMux] explains how patterns are matched.
func (o openTSG) HandleFunc(wType string, handler HandlerFunc) {
	// set up router here

	o.Handle(wType, []byte("{}"), handler)

}

func (o openTSG) Handle(wType string, schema []byte, handler Handler) {

	if _, ok := o.handlers[wType]; ok {
		panic(fmt.Sprintf("The widget type %s has already been declared", wType))
	}

	// do some checking for invalid characters, if there
	// are any

	o.handlers[wType] = hand{schema: schema, handler: handler}
}

func (o *openTSG) Use(middlewares ...func(Handler) Handler) {
	o.middlewares = append(o.middlewares, middlewares...)
}

type Request struct {
	// For http handlers etc
	RawWidgetYAML json.RawMessage
	// the properties of the patch to be made
	PatchProperties PatchProperties
	FrameProperties FrameProperties

	// Helper functions that communicate with the engine
	// exported as methods
	// these are not exported as a json for http requests
	// the context is passed to widgets for this
	// offer a default for the text box search

	searchWithCredentials func(URI string) ([]byte, error)
	// generate an image of smaller dimensions
	// Not really needed? Look through
	// IMG(image.Rect)Draw.Image
	// GetWidgetMetadata()
}

func (r Request) SearchWithCredentials(URI string) ([]byte, error) {
	if r.searchWithCredentials == nil {
		return core.GetWebBytes(nil, URI)
	}
	return r.searchWithCredentials(URI)
}

func GenerateSubImage(baseImg draw.Image, bounds image.Rectangle) draw.Image {
	return nil
}

type PatchProperties struct {
	WidgetType   string
	WidgetFullID string
	// only do max xy as min is always 0
	Dimensions  image.Rectangle
	TSGLocation image.Point
	// Padding and margins?
	Geomtetry   []gridgen.Segmenter
	ColourSpace colour.ColorSpace
}

type FrameProperties struct {
	FrameNumber int
	WorkingDir  string
}

type Response interface {
	// Write a response to signal
	// the end of the widget and to handle any errors.
	// This is not vital
	Write(status int, message string)
	// Write(logLevel slog.Level ,status int, message string)

	// just draw the image as standard
	// import it as our own in case it needs
	// to be extended later
	draw.Image
}

type response struct {
	baseImg draw.Image
	status  int
	message string
}

func (r *response) Write(status int, message string) {
	// nothing is written at the moment
	r.status = status
	r.message = message
}

func (r *response) At(x int, y int) color.Color {
	return r.baseImg.At(x, y)
}

func (r *response) Bounds() image.Rectangle {
	return r.baseImg.Bounds()
}
func (r *response) ColorModel() color.Model {
	return r.baseImg.ColorModel()
}

func (r *response) Set(x, y int, c color.Color) {
	r.baseImg.Set(x, y, c)
}

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

func (tsg *openTSG) logErrors(code, frameNumer int, errors ...error) {
	errHan := HandlerFunc(func(resp Response, req *Request) {
		resp.Write(code, string(req.RawWidgetYAML))
	})
	errs := chain(tsg.middlewares, errHan)
	// call all errors so they are just logged
	for _, err := range errors {
		errs.Handle(&response{}, &Request{RawWidgetYAML: json.RawMessage(err.Error()),
			PatchProperties: PatchProperties{WidgetFullID: "core.tsg"},
			FrameProperties: FrameProperties{FrameNumber: frameNumer},
		})
	}
}

// Draw generates the images for each array section of the json array and applies it to the test card grid.
func (tsg *openTSG) Run(mnt string) {
	imageNo := tsg.framcount

	// wait for every frame to run before exiting the lopp
	var wg sync.WaitGroup
	wg.Add(tsg.framcount)

	// hookdata is a large map that contains all the metadata across the run.
	var locker sync.Mutex
	hookdata := syncmap{&locker, make(map[string]any)}

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

			monit := monitor{frameNo: frameNo}
			genMeasure := time.Now()
			saveTime := int64(0)
			// new log here for each frame

			// defer the progress bar message to use the values at the end of the "function"
			// the idea is for them to auto update
			defer func() {
				tsg.logErrors(200, frameNo,
					fmt.Errorf("generating frame %v/%v, gen: %v ms, save: %sms, errors:%v", frameNo, imageNo-1,
						microToMili(int64(time.Since(genMeasure).Microseconds())), microToMili(saveTime), monit.ErrorCount),
				)
				// add the log to the cache channel

			}()

			// update metadata to be included in the frame context
			frameConfigCont, errs := core.FrameWidgetsGenerator(tsg.internal, frameNo)

			// this is important for showing missed widget updates
			// log the errors
			if len(errs) > 0 {
				tsg.logErrors(404, frameNo, errs...)
				monit.incrementError(len(errs))
			}

			frameContext := widgethandler.MetaDataInit(frameConfigCont)
			errs = canvaswidget.LoopInit(frameContext)

			if len(errs) > 0 {
				// log.Fatal
				tsg.logErrors(500, frameNo, errs...)
				monit.incrementError(len(errs))
				// frameWait.Done() //the frame weight is returned when the programs exit, or the frame has been generated
				return // continue // skip to the next frame number
			}
			// generate the canvas of type image.Image
			canvas, err := gridgen.GridGen(frameContext)
			if err != nil {
				tsg.logErrors(500, frameNo, err)
				monit.incrementError(1)
				// frameWait.Done()
				return // continue // skip to the next frame number
			}

			// generate all the widgets
			tsg.widgetHandle(frameContext, canvas, &monit)
			// frameWait.Done()

			// get the metadata and add it onto the map for this frame
			md, _ := metaHook(canvas, frameContext)
			if len(md) != 0 { // only save if there actually is metadata
				i4 := intToLength(frameNo, 4)
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
func (tsg *openTSG) canvasSave2(canvas draw.Image, filename []string, bitdeph int, mnt string, monit *monitor) {
	for _, name := range filename {
		truepath, err := filepath.Abs(filepath.Join(mnt, name))
		if err != nil {
			monit.incrementError(1)
			tsg.logErrors(700, monit.frameNo, err)

			continue
		}
		err = tsg.encodeFrame(truepath, canvas, bitdeph)
		if err != nil {
			monit.incrementError(1)
			tsg.logErrors(700, monit.frameNo, err)
		}
	}
}

type monitor struct {
	frameNo    int
	ErrorCount int
	sync.Mutex
}

func (m *monitor) incrementError(count int) {
	m.Lock()
	m.ErrorCount += count
	m.Unlock()
}

// // update widgetHandle to make the choices for me
func (tsg *openTSG) widgetHandle(c *context.Context, canvas draw.Image, monit *monitor) {

	// set up the core context functions
	allWidgets := widgets.ExtractAllWidgets(c)
	// add the validator last
	lineErrs := core.GetJSONLines(*c)
	webSearch := func(URI string) ([]byte, error) {
		return core.GetWebBytes(c, URI)
	}

	// get the widgtes to be used
	allWidgetsArr := make([]core.AliasIdentity, len(allWidgets))
	for alias := range allWidgets {

		allWidgetsArr[alias.ZPos] = alias

	}

	// set up the properties for all requests
	fp := FrameProperties{WorkingDir: core.GetDir(*c), FrameNumber: monit.frameNo}

	// sync tools for running the widgets async
	runPool := Pool{AvailableMemeory: tsg.ruunerConf.RunnerCount}
	// wg for each widget
	var wg sync.WaitGroup
	wg.Add(len(allWidgets))
	// ensure z order
	// prevent race conditions writing to the canvas
	zpos := 0
	zPos := &zpos
	var zPosLock sync.Mutex
	var canvasLock sync.Mutex

	for i := 0; i < len(allWidgets); i++ {

		// get a runner to run the widget
		runner, available := runPool.GetRunner()
		for !available {

			time.Sleep(10 * time.Millisecond)
			runner, available = runPool.GetRunner()
		}

		// run the widget async
		go func() {
			position := i
			defer runPool.PutRunner(runner)
			defer wg.Done()

			widg := allWidgets[allWidgetsArr[i]]
			widgProps := allWidgetsArr[i]

			if widgProps.WType != "builtin.canvasoptions" && widgProps.WType != "" {

				handlers, handlerExists := tsg.handlers[allWidgetsArr[i].WType]
				// make a function so the handler is returned
				// @TODO skip the handler and come back to it later

				var Han Handler
				var resp response
				var req Request
				var gridcanvas, mask draw.Image
				var imgLocation image.Point

				// run a set up function that can return early
				// to make the handler just spit out the error
				func() {
					// ensure the chain is always kept
					defer func() {
						Han = chain(tsg.middlewares, Han)
					}()

					if !handlerExists {
						Han = GenErrorHandler(400,
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

					gridcanvas, imgLocation, mask, err = gridgen.GridSquareLocatorAndGenerator(widgProps.Location, widgProps.GridAlias, c)
					// when the function am error is returned,
					// the function just becomes return an error
					if err != nil {
						Han = GenErrorHandler(400, err.Error())
						return
					}

					flats, err := gridgen.GetGridGeometry(c, widgProps.Location)
					if err != nil {
						Han = GenErrorHandler(400, err.Error())
						return
					}

					// set up the requests
					// and chain the middleware for the handler

					pp := PatchProperties{WidgetType: widgProps.WType,
						WidgetFullID: widgProps.FullName,
						Dimensions:   gridcanvas.Bounds(),
						TSGLocation:  imgLocation, Geomtetry: flats,
						ColourSpace: widgProps.ColourSpace}
					//	Han, err := Unmarshal(handlers.handler)(widg)
					resp = response{baseImg: gridcanvas}
					req = Request{FrameProperties: fp, RawWidgetYAML: widg,
						searchWithCredentials: webSearch, PatchProperties: pp,
					}

					// chain that middleware at the last second?
					validatorMid := jSONValidator(lineErrs, handlers.schema, widgProps.FullName)
					Han = chain([]func(Handler) Handler{validatorMid}, Han)

				}()

				// RUN the widget
				Han.Handle(&resp, &req)

				// wait until it is the widgets turn
				zPosLock.Lock()
				widgePos := *zPos
				zPosLock.Unlock()
				for widgePos != position {
					time.Sleep(time.Millisecond * 10)
					zPosLock.Lock()
					widgePos = *zPos
					zPosLock.Unlock()
				}

				// only draw the image if
				// no errors occurred running the handler
				if resp.status == 200 {
					canvasLock.Lock()
					colour.DrawMask(canvas, gridcanvas.Bounds().Add(imgLocation), gridcanvas, image.Point{}, mask, image.Point{}, draw.Over)
					canvasLock.Unlock()
				} else {
					// error of some sort from somewhere
					monit.incrementError(1)
				}
				// calculate some response stuff
			} else {
				// don't skip the widget stuff
				zPosLock.Lock()
				widgePos := *zPos
				zPosLock.Unlock()
				for widgePos != position {
					time.Sleep(time.Millisecond * 10)
					zPosLock.Lock()
					widgePos = *zPos
					zPosLock.Unlock()
				}
			}

			zPosLock.Lock()
			// update zpos regardless
			*zPos++
			zPosLock.Unlock()
		}()
	}

	wg.Wait()
}

func GenErrorHandler(code int, errMessage string) Handler {
	return HandlerFunc(func(r Response, _ *Request) {
		r.Write(code, errMessage)
	})
}

type Pool struct {
	// keep at 1 at the moment
	AvailableMemeory int
	sync.Mutex
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
	/*
		lock try and get a runner

	*/

	return
}

func (p *Pool) PutRunner(run poolRunner) {

	p.Lock()
	defer p.Unlock()
	p.AvailableMemeory += run.memory
	/*
		lock etc
	*/
}

type poolRunner struct {
	memory int
}

/*

sync pool, figure something out

*/

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
