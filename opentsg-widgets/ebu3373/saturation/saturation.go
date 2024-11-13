// package saturation generates the ebu3373 saturation boxes
package saturation

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
)

const (
	WidgetType = "builtin.ebu3373/saturation"
)

func SatGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := widgethandler.GenConf[Config]{Debug: debug, Schema: Schema, WidgetType: WidgetType}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

var (
	grey = colour.CNRGBA64{R: 26496, G: 26496, B: 26496, A: 0xffff}
)

var (
	satRed0  = colour.CNRGBA64{R: 793 << 6, G: 793 << 6, B: 793 << 6, A: 0xffff}
	satRed1  = colour.CNRGBA64{R: 815 << 6, G: 779 << 6, B: 763 << 6, A: 0xffff}
	satRed2  = colour.CNRGBA64{R: 834 << 6, G: 762 << 6, B: 730 << 6, A: 0xffff}
	satRed3  = colour.CNRGBA64{R: 851 << 6, G: 742 << 6, B: 695 << 6, A: 0xffff}
	satRed4  = colour.CNRGBA64{R: 866 << 6, G: 718 << 6, B: 654 << 6, A: 0xffff}
	satRed5  = colour.CNRGBA64{R: 881 << 6, G: 689 << 6, B: 607 << 6, A: 0xffff}
	satRed6  = colour.CNRGBA64{R: 894 << 6, G: 651 << 6, B: 549 << 6, A: 0xffff}
	satRed7  = colour.CNRGBA64{R: 907 << 6, G: 600 << 6, B: 473 << 6, A: 0xffff}
	satRed8  = colour.CNRGBA64{R: 919 << 6, G: 522 << 6, B: 386 << 6, A: 0xffff}
	satRed9  = colour.CNRGBA64{R: 930 << 6, G: 390 << 6, B: 283 << 6, A: 0xffff}
	satRed10 = colour.CNRGBA64{R: 940 << 6, G: 64 << 6, B: 64 << 6, A: 0xffff}
)

var (
	satGreen0  = colour.CNRGBA64{R: 809 << 6, G: 809 << 6, B: 809 << 6, A: 0xffff}
	satGreen1  = colour.CNRGBA64{R: 791 << 6, G: 830 << 6, B: 783 << 6, A: 0xffff}
	satGreen2  = colour.CNRGBA64{R: 770 << 6, G: 848 << 6, B: 756 << 6, A: 0xffff}
	satGreen3  = colour.CNRGBA64{R: 747 << 6, G: 863 << 6, B: 726 << 6, A: 0xffff}
	satGreen4  = colour.CNRGBA64{R: 720 << 6, G: 878 << 6, B: 691 << 6, A: 0xffff}
	satGreen5  = colour.CNRGBA64{R: 688 << 6, G: 890 << 6, B: 651 << 6, A: 0xffff}
	satGreen6  = colour.CNRGBA64{R: 648 << 6, G: 902 << 6, B: 601 << 6, A: 0xffff}
	satGreen7  = colour.CNRGBA64{R: 594 << 6, G: 913 << 6, B: 536 << 6, A: 0xffff}
	satGreen8  = colour.CNRGBA64{R: 511 << 6, G: 922 << 6, B: 443 << 6, A: 0xffff}
	satGreen9  = colour.CNRGBA64{R: 380 << 6, G: 931 << 6, B: 327 << 6, A: 0xffff}
	satGreen10 = colour.CNRGBA64{R: 64 << 6, G: 940 << 6, B: 64 << 6, A: 0xffff}
)

var (
	satBlue0  = colour.CNRGBA64{R: 579 << 6, G: 578 << 6, B: 578 << 6, A: 0xffff}
	satBlue1  = colour.CNRGBA64{R: 552 << 6, G: 584 << 6, B: 631 << 6, A: 0xffff}
	satBlue2  = colour.CNRGBA64{R: 520 << 6, G: 586 << 6, B: 677 << 6, A: 0xffff}
	satBlue3  = colour.CNRGBA64{R: 484 << 6, G: 584 << 6, B: 717 << 6, A: 0xffff}
	satBlue4  = colour.CNRGBA64{R: 445 << 6, G: 578 << 6, B: 755 << 6, A: 0xffff}
	satBlue5  = colour.CNRGBA64{R: 403 << 6, G: 564 << 6, B: 790 << 6, A: 0xffff}
	satBlue6  = colour.CNRGBA64{R: 360 << 6, G: 542 << 6, B: 823 << 6, A: 0xffff}
	satBlue7  = colour.CNRGBA64{R: 312 << 6, G: 504 << 6, B: 854 << 6, A: 0xffff}
	satBlue8  = colour.CNRGBA64{R: 260 << 6, G: 443 << 6, B: 884 << 6, A: 0xffff}
	satBlue9  = colour.CNRGBA64{R: 197 << 6, G: 347 << 6, B: 912 << 6, A: 0xffff}
	satBlue10 = colour.CNRGBA64{R: 64 << 6, G: 64 << 6, B: 940 << 6, A: 0xffff}
)

func (s Config) Generate(canvas draw.Image, opt ...any) error {
	b := canvas.Bounds().Max

	// Use a map to match the keys up etc
	reds := []colour.CNRGBA64{satRed0, satRed1, satRed2, satRed3, satRed4, satRed5, satRed6, satRed7, satRed8, satRed9, satRed10}
	greens := []colour.CNRGBA64{satGreen0, satGreen1, satGreen2, satGreen3, satGreen4, satGreen5, satGreen6, satGreen7, satGreen8, satGreen9, satGreen10}
	blues := []colour.CNRGBA64{satBlue0, satBlue1, satBlue2, satBlue3, satBlue4, satBlue5, satBlue6, satBlue7, satBlue8, satBlue9, satBlue10}
	colours := make(map[string][]colour.CNRGBA64)
	cs := [][]colour.CNRGBA64{reds, greens, blues}
	names := []string{"red", "green", "blue"}
	for i, c := range cs {
		colours[names[i]] = c
	}
	var inputC []string
	// Assign the basic colour order if none are chosen
	if len(s.Colours) == 0 {
		inputC = names
	} else {
		inputC = s.Colours
	}

	wScale := (float64(b.X) / 2330.0)
	hScale := (float64(b.Y) / 600.0)

	height := 0.0
	// Scale the height offset to the number of colours called (max 3)
	hOff := hScale * 200 * (3.0 / float64(len(inputC)))

	colour.Draw(canvas, canvas.Bounds(), &image.Uniform{&grey}, image.Point{}, draw.Src)

	wOff := wScale * 200 // 200 is the width at 3840
	for _, cname := range inputC {
		fillColour := colours[cname]
		if fillColour == nil {
			// Blow the doors of etc
			return fmt.Errorf("TEST the colour %v is not an available colour", cname)
		}
		// Draw the narrow box
		width := wScale * 100
		start := fillColour[0]
		start.UpdateColorSpace(s.ColourSpace)
		colour.Draw(canvas, image.Rect(int(0), int(height), int(width), int(height+hOff)), &image.Uniform{&start}, image.Point{}, draw.Over)
		// Draw the smaller boxes at a smaller fill
		for _, c := range fillColour[1:] {
			offx := 50 * wScale
			offy := 50 * hScale
			fill := c
			fill.UpdateColorSpace(s.ColourSpace)
			colour.Draw(canvas, image.Rect(int(width+offx), int(height+offy), int(width+wOff-offx), int(height+hOff-offy)), &image.Uniform{&fill}, image.Point{}, draw.Over)
			width += wOff
		}
		// Draw the final large box
		end := fillColour[len(fillColour)-1]
		end.UpdateColorSpace(s.ColourSpace)
		colour.Draw(canvas, image.Rect(int(width), int(height), b.X, int(height+hOff)), &image.Uniform{&end}, image.Point{}, draw.Over)
		height += hOff
	}

	return nil
}

func (s Config) Handle(resp tsg.Response, req *tsg.Request) {
	b := resp.BaseImage().Bounds().Max

	// Use a map to match the keys up etc
	reds := []colour.CNRGBA64{satRed0, satRed1, satRed2, satRed3, satRed4, satRed5, satRed6, satRed7, satRed8, satRed9, satRed10}
	greens := []colour.CNRGBA64{satGreen0, satGreen1, satGreen2, satGreen3, satGreen4, satGreen5, satGreen6, satGreen7, satGreen8, satGreen9, satGreen10}
	blues := []colour.CNRGBA64{satBlue0, satBlue1, satBlue2, satBlue3, satBlue4, satBlue5, satBlue6, satBlue7, satBlue8, satBlue9, satBlue10}
	colours := make(map[string][]colour.CNRGBA64)
	cs := [][]colour.CNRGBA64{reds, greens, blues}
	names := []string{"red", "green", "blue"}
	for i, c := range cs {
		colours[names[i]] = c
	}
	var inputC []string
	// Assign the basic colour order if none are chosen
	if len(s.Colours) == 0 {
		inputC = names
	} else {
		inputC = s.Colours
	}

	wScale := (float64(b.X) / 2330.0)
	hScale := (float64(b.Y) / 600.0)

	height := 0.0
	// Scale the height offset to the number of colours called (max 3)
	hOff := hScale * 200 * (3.0 / float64(len(inputC)))

	colour.Draw(resp.BaseImage(), resp.BaseImage().Bounds(), &image.Uniform{&grey}, image.Point{}, draw.Src)

	wOff := wScale * 200 // 200 is the width at 3840
	for _, cname := range inputC {
		fillColour := colours[cname]
		if fillColour == nil {
			// Blow the doors of etc
			resp.Write(tsg.WidgetError, fmt.Sprintf("The colour %v is not an available colour", cname))
			return
		}
		// Draw the narrow box
		width := wScale * 100
		start := fillColour[0]
		start.UpdateColorSpace(s.ColourSpace)
		colour.Draw(resp.BaseImage(), image.Rect(int(0), int(height), int(width), int(height+hOff)), &image.Uniform{&start}, image.Point{}, draw.Over)
		// Draw the smaller boxes at a smaller fill
		for _, c := range fillColour[1:] {
			offx := 50 * wScale
			offy := 50 * hScale
			fill := c
			fill.UpdateColorSpace(s.ColourSpace)
			colour.Draw(resp.BaseImage(), image.Rect(int(width+offx), int(height+offy), int(width+wOff-offx), int(height+hOff-offy)), &image.Uniform{&fill}, image.Point{}, draw.Over)
			width += wOff
		}
		// Draw the final large box
		end := fillColour[len(fillColour)-1]
		end.UpdateColorSpace(s.ColourSpace)
		colour.Draw(resp.BaseImage(), image.Rect(int(width), int(height), b.X, int(height+hOff)), &image.Uniform{&end}, image.Point{}, draw.Over)
		height += hOff
	}

	resp.Write(tsg.WidgetSuccess, "success")
}
