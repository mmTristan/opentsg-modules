// Package qrgen generates a qr code based on user string and places it on the graph, this is the last item to be added
package qrgen

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"math"
	"sync"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mmTristan/opentsg-modules/opentsg-core/errHandle"
	"github.com/mmTristan/opentsg-modules/opentsg-core/widgethandler"
)

const (
	widgetType = "builtin.qrcode"
)

func QrGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := widgethandler.GenConf[qrcodeJSON]{Debug: debug, Schema: schemaInit, WidgetType: widgetType, ExtraOpt: []any{c}}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

// var extract = widgethandler.Extract

func (qrC qrcodeJSON) Generate(canvas draw.Image, opt ...any) error {
	message := qrC.Code
	if message == "" {
		// Return but don't fill up the stdout with errors
		return nil
	}
	/*
		TODO: utilise this information for metadata in the barcode
			if qrC.Query != nil {
				// Do some more metadata extraction
				for _, q := range *qrC.Query {
					fmt.Println(q)
					fmt.Println(extract(opt[0].(*context.Context), q.Target, q.Keys...))
				}
			}
	*/

	code, err := qr.Encode(message, qr.H, qr.Auto)
	if err != nil {
		return fmt.Errorf("0131 %v", err)
	}

	b := canvas.Bounds().Max
	if qrC.Size != nil {
		width, height := qrC.Size.Width, qrC.Size.Height
		if width != 0 && height != 0 {
			w, h := (width/100)*float64(b.X), (height/100)*float64(b.Y)
			code, err = barcode.Scale(code, int(w), int(h))
			if err != nil {
				return fmt.Errorf("0132 %v", err)
			}
		}
	}

	var x, y int
	if qrC.Imgpos != nil { // Scale x and y as a percentage
		x = int(math.Floor((qrC.Imgpos.X / 100) * float64(b.X)))
		y = int(math.Floor((qrC.Imgpos.Y / 100) * float64(b.Y)))
	}

	if x > (b.X - code.Bounds().Max.X) {
		return fmt.Errorf("0133 the x position %v is greater than the x boundary of %v", x, canvas.Bounds().Max.X)
	} else if y > b.Y-code.Bounds().Max.Y {
		return fmt.Errorf("0133 the y position %v is greater than the y boundary of %v", y, canvas.Bounds().Max.Y)
	}
	// draw qr code as a mid point, or make colour space agnostic
	colour.Draw(canvas, canvas.Bounds().Add(image.Point{x, y}), code, image.Point{}, draw.Over)

	return nil
}
