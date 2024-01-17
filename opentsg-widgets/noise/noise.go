// Package noise generates images of noise
package noise

import (
	"context"
	"fmt"
	"image/draw"
	"math/rand"
	"sync"
	"time"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mmTristan/opentsg-modules/opentsg-core/errHandle"
	"github.com/mmTristan/opentsg-modules/opentsg-core/widgethandler"
)

const (
	widgetType = "builtin.noise"
)

const (
	whiteNoise = "white noise"
)

// NGenerator generates images of noise
func NGenerator(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := widgethandler.GenConf[noiseJSON]{Debug: debug, Schema: schemaInit, WidgetType: widgetType}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

var randnum = randSeed

func randSeed() int64 {
	return time.Now().Unix()
}

func (n noiseJSON) Generate(canvas draw.Image, opt ...any) error {

	// Have a seed variable tht is taken out for testing purposes
	random := rand.New(rand.NewSource(randnum()))

	var max int
	if n.Maximum != 0 {
		max = n.Maximum
	} else {
		// Revert to the default
		max = 4095
	}
	min := n.Minimum

	if max < min {
		return fmt.Errorf("0141 The minimum noise value %v is greater than the maximum noise value %v", min, max)
	}

	if n.NoiseType == whiteNoise { // upgrade to switch statement when more types come in
		whitenoise(random, n.ColourSpace, canvas, min, max)
	}

	return nil
}

func whitenoise(random *rand.Rand, cspace colour.ColorSpace, canvas draw.Image, min, max int) {
	b := canvas.Bounds().Max
	for y := 0; y < b.Y; y++ {
		for x := 0; x < b.X; x++ {
			colourPos := uint16(random.Intn(max-min)+min) << 4
			// Fill := color.NRGBA64{colourPos << 4, colourPos << 4, colourPos << 4, uint16(0xffff)}

			canvas.Set(x, y, &colour.CNRGBA64{R: colourPos, G: colourPos, B: colourPos, A: 0xffff, ColorSpace: cspace})
		}
	}
}
