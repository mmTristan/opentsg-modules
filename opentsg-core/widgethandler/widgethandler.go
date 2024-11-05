// package widethandler is used for generically running the widgets in the order they were assigned
package widgethandler

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"strings"
	"sync"
	"time"

	"github.com/mrmxf/opentsg-modules/opentsg-core/canvaswidget"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/widgets"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"gopkg.in/yaml.v3"
)

// Generator contains the method for running widgets to generate segments of the test chart.
type Generator interface {
	Generate(draw.Image, ...any) error
	// Loc returns the location of the grid for  gridgen.ParamToCanvas
	Location() string
	// Alias returns the alias of the grid for  gridgen.ParamToCanvas
	Alias() string
}

// GenConf is the input struct of information for the WidgetRunner function
type GenConf[T Generator] struct {
	Debug      bool
	Schema     []byte
	WidgetType string
	ExtraOpt   []any
}

// WidgetRunner is the order in which the images are made and added to the canvas, this is a generic function so that
// each widget can call it and run.
// widgets should follow the pattern of
//
//	func WidgetRunner(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *log.Logger) {
//		defer wg.Done()
//		conf := generator.GenConf[widgetstruct]{Debug: debug, Schema: schemaFac, WidgetType: "widgetname"}
//		generator.WidgetRunner(canvasChan, conf, c, logs, wgc)
//	}
/*
// Generator has the following interface

type Generator interface {
	Generate(draw.Image, ...any) error
	//Loc returns the location of the grid for gridgen.ParamToCanvas
	Location() string
	//Alias returns the alias of the grid for gridgen.ParamToCanvas
	Alias() string
}
*/
func WidgetRunner[T Generator](canvasChan chan draw.Image, g GenConf[T], c *context.Context, logs *errhandle.Logger, wgc *sync.WaitGroup) {
	// get the update map to run

	errorCode := strings.ToUpper(g.WidgetType)
	// change config factory runmap to out base map alias identity with func T and sidecar of information - parents etc?
	runMap, err := widgets.ExtractWidgetStructs[T](g.WidgetType, g.Schema, c)

	putErr := put(runMap, c) // put metadata info in and background contexts to run the generator function

	wgc.Done() // wait for all other widgets to add their metadata
	wgc.Wait()

	if putErr != nil {
		logs.PrintErrorMessage("E_"+errorCode+"_INIT_", putErr, g.Debug)
	}

	// print any error message before you can return
	if len(err) > 0 {
		for _, e := range err {
			logs.PrintErrorMessage("E_"+errorCode+"_INIT_", e, g.Debug)
		}
	}

	// get the widget runorder
	runOrder := getOrder(runMap)
	if len(runOrder) == 0 {
		return
	}

	/*	if len(err) > 0 {

		for _, e := range err {
			logs.PrintErrorMessage("E_"+errorCode+"_INIT_", e, g.Debug)
		}
		/*run as if the images had been made so the zpos does not fall out of sync
		and we are stuck with an endless cycle
		count := 0 // these runorders no longer run for errored messages as they are not assigned
		for count < len(runOrder) {
			z := (*c).Value(zKey).(*int)
			if *z == runOrder[count].ZPos {

				count++
				*z++

			}
		}

		return
	} */

	// wait until the first position is ready to run to stop overloads
	for {
		z := (*c).Value(zKey).(*int)
		if *z == runOrder[0].ZPos {
			break
		} else { // add a panic here after 10 minutes or so
			time.Sleep(10 * time.Millisecond)
		}
	}

	// go for a new method
	// where the image is generated and then blocks until its zorder is found
	for _, key := range runOrder {
		var generated bool
		v := runMap[key]
		k := key.FullName
		gridcanvas, imgLocation, mask, err := gridgen.GridSquareLocatorAndGenerator(v.Location(), v.Alias(), c)

		if err != nil {
			logs.PrintErrorMessage(fmt.Sprintf("E_%v_%v_", errorCode, k), err, g.Debug)

		} else {
			// amending to remove having a 16 image in storage
			// generate the image
			if err := v.Generate(gridcanvas, g.ExtraOpt...); err != nil {
				logs.PrintErrorMessage(fmt.Sprintf("E_%v_%v_", errorCode, k), err, g.Debug)

			} else {
				// only generate if no errors were found
				generated = true
			}
		}
		zfound := false
		for !zfound {

			z := (*c).Value(zKey).(*int)
			if *z == key.ZPos {
				if generated {
					// extract base from channel before readding and preventing race conditions
					canvas := <-canvasChan
					colour.DrawMask(canvas, gridcanvas.Bounds().Add(imgLocation), gridcanvas, image.Point{}, mask, image.Point{}, draw.Over)
					// draw.DrawMask(canvas, canvas.Bounds(), add.img, add.location, add.mask, add.location, draw.Over)
					canvasChan <- canvas
				}
				// counter and z have two different scopes

				*z++
				zfound = true
			} else { // add a panic here after 10 minutes or so
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	/*
		// run the image generation concurrently to adding it
		go func() {
			// After the hook check generate the image
			for i, key := range runOrder {

				v := runMap[key]
				k := key.Alias
				Img, imgLocation, mask, err := gridgen.GridSquareLocatorAndGenerator(v.Location(), v.Alias(), c)

				if err != nil {
					logs.PrintErrorMessage(fmt.Sprintf("E_%v_%v_", errorCode, k), err, g.Debug)
					additions[i] = adder{zPos: key.ZPos, run: true}

					continue
				}
				// amending to remove having a 16 image in storage
				// generate the image
				if err := v.Generate(Img, g.ExtraOpt...); err != nil {
					logs.PrintErrorMessage(fmt.Sprintf("E_%v_%v_", errorCode, k), err, g.Debug)
					additions[i] = adder{zPos: key.ZPos, run: true}
				} else {
					// add to the array all the additions to be made
					additions[i] = adder{img: Img, mask: mask, location: imgLocation, zPos: key.ZPos, add: true, run: true}

				}
			}
		}()

		// basic zposition counter is looped across all the widgets
		counter := 0

		for counter < len(additions) {
			// z := (*c).Value(zKey).(*int)
			// fmt.Println(*z, counter, errorCode, additions)
			if additions[counter].run {
				add := additions[counter]
				z := (*c).Value(zKey).(*int)
				// fmt.Println(*z, errorCode)
				if *z == add.zPos {
					if add.add {
						// extract base from channel before readding and preventing race conditions
						canvas := <-canvasChan
						draw.DrawMask(canvas, add.img.Bounds().Add(add.location), add.img, image.Point{}, add.mask, image.Point{}, draw.Over)
						//draw.DrawMask(canvas, canvas.Bounds(), add.img, add.location, add.mask, add.location, draw.Over)
						canvasChan <- canvas
					}
					// counter and z have two different scopes
					counter++
					*z++
				} else {
					// sleep until it reaches the z position and the counter has the vlaue
					time.Sleep(10 * time.Millisecond)
				}
			} else { // add a panic here after 10 minutes or so
				time.Sleep(10 * time.Millisecond)
			}

		}*/

}

func getOrder[A any](runMap map[core.AliasIdentity]A) []core.AliasIdentity {
	runOrder := make([]core.AliasIdentity, len(runMap))
	// figure out an order here
	i := 0
	for key := range runMap {
		runOrder[i] = key
		i++
	}

	// sort everything in order
	for i := range runOrder {
		for j := 0; j < len(runOrder)-1; j++ {
			if runOrder[i].ZPos < runOrder[j].ZPos {
				runOrder[j], runOrder[i] = runOrder[i], runOrder[j]
			}
		}
	}

	return runOrder
}

////////////////////////
// metadata handling //
//////////////////////

// unique name for this package so information can't be extracted else where
const (
	metakey contextKey = "metadata"
	zKey    contextKey = "pointer to the z position of all the widgets"
)

type contextKey string

// metadata is a map with a mutex to prevent concurrent read write to maps
type metadata struct {
	data map[string]map[string]interface{}
	mu   *sync.Mutex
}

// Put adds the information used to generate widget to the metadata
// is unexported to keep the information immutable
func put[T any](toSave map[core.AliasIdentity]T, c *context.Context) error {

	// prevent concurrent writes
	imageGeneration := (*c).Value(metakey).(metadata)
	imageGeneration.mu.Lock()
	defer imageGeneration.mu.Unlock()
	// map[string]map[string]map[string]interface{})

	// breakdown the map and add it to the image generation
	for k, v := range toSave { // if empty than it's skipped
		readForm := make(map[string]interface{})
		b, err := yaml.Marshal(v) // reset it from struct type to map[string]interface{}
		if err != nil {

			return fmt.Errorf("0201 Error inserting metadata %v", err)
		}

		err = yaml.Unmarshal(b, &readForm)

		if err != nil {

			return fmt.Errorf("0201 Error inserting metadata %v", err)
		}
		imageGeneration.data[k.FullName] = readForm
	}
	// imageGeneration.data[widget] = alias

	return nil
}

// metaDataInit adds the metadata to the global context for that frame. This needs to be called
// before the widgets are run so that metadata can be stored.
func MetaDataInit(c context.Context) *context.Context {
	// MD is the metadata context of widget - alias - json(map[string] interface{})
	md := metadata{make(map[string]map[string]interface{}), &sync.Mutex{}}
	valC := context.WithValue(c, metakey, md)
	z := 0

	valC = context.WithValue(valC, zKey, &z)

	return &valC
}

// Extract searches the metadata for an alias and then recursivley searches through the keys to find nested values.
// The keys are treated as a dotpath and will follow the path of any previous keys.
// if the alias is "" then all the metadata is returned.
func Extract(c *context.Context, alias string, keys ...string) interface{} {
	// order is widget type, alias map of json information
	imageGeneration := (*c).Value(metakey).(metadata)
	// map[string]map[string]map[string]interface{})
	if alias == "" {

		return imageGeneration.data
	}
	start := imageGeneration.data[alias]
	if len(keys) != 0 {

		return mapToFace(start, keys...)
	}

	return start
}

func mapToFace(find map[string]any, keys ...string) interface{} {

	fmt.Println(keys)
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

type widgetProperties struct {
	Identity, Alias string
}

// Extract searches the data for the widget types and then returns an array of the aliases with their idnetity.
// If not types are given then everything widget and their alias is returned
func ExtractWidget(c *context.Context, types ...string) []widgetProperties {
	imageGeneration := (*c).Value(metakey).(map[string]map[string]map[string]interface{})
	var result []widgetProperties

	var all bool
	if len(types) == 0 { // then return all
		all = true
	}
	for _, target := range types {
		// loop through types and check if it matches
		for wType, alias := range imageGeneration { // then extract all the widgets
			if wType == target || all {
				for key := range alias {
					result = append(result, widgetProperties{wType, key})
				}
			}
		}
	}

	return result
}

///// CANVASES /////
// this is a canvas dummy that is being used within the system to utilise the zposition assigned to it

// MockCanvasGen mocks the canvas widget, so that the zposition is updated and the canvas information is added to the metadata
func MockCanvasGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := GenConf[canvaswidget.ConfigVals]{Debug: debug, Schema: canvaswidget.GetCanvasSchema(), WidgetType: "builtin.canvasoptions"}
	WidgetRunner(canvasChan, conf, c, logs, wgc) // update this to pass an error which is then formatted afterwards
}

// MockedMissedGen runs as a pseudo generator function to check for any missed widget names and preserve the z order of the frame.
// It prevents z orders not being updated for unassigend widgets and the resulting time outs, as well as
// informing the user of any unintentional missed widget.
func MockMissedGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()

	wgc.Done()
	wgc.Wait()

	missed := widgets.MissingWidgetCheck(*c)
	// if nothing has been missed return early
	if len(missed) == 0 {

		return
	}

	// get the order of missed names
	runOrder := getOrder(missed)
	missedNames := make([]string, len(runOrder))

	count := 0
	// loop through saying they are completed to preserve the z order
	for count < len(runOrder) {
		z := (*c).Value(zKey).(*int)
		if *z == runOrder[count].ZPos {
			missedNames[count] = runOrder[count].FullName
			count++
			*z++
		}
	}
	e := fmt.Errorf("the following name(s) were not assigned a widget: %v", missedNames)
	logs.PrintErrorMessage("W_CORE_opentsg_", e, debug)
}
