// Package addimage allows images to be uploaded and added to the canvas.
package addimage

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"

	"github.com/nfnt/resize"
	"golang.org/x/image/tiff"
)

const (
	WidgetType = "builtin.addimage"
)

// ImageGen opens an image and places it at the user specife grid location,
// the image may be resized if the sizes do not match up.
// Only 8/16 bit PNG and TIFF files are valid file to be uploaded.
func ImageGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := widgethandler.GenConf[Config]{Debug: debug, Schema: Schema, WidgetType: WidgetType, ExtraOpt: []any{c}}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc)
}

func (c Config) Handle(resp tsg.Response, req *tsg.Request) {
	filename := c.Image
	if filename == "" {
		resp.Write(tsg.WidgetError, fmt.Sprintf("0161 No image declared"))
		return
	}

	wDir := req.FrameProperties.WorkingDir
	// Just check if it's a website first
	webBytes, errOpen := req.SearchWithCredentials(filename)
	var newImage image.Image
	var err error
	var depth int
	// Open a local file next if not
	if errOpen != nil {

		file, errOpen := os.Open(filepath.Join(wDir, filename))
		if errOpen != nil {
			resp.Write(tsg.WidgetError, errOpen.Error())
			return
		}
		newImage, depth, err = fToImg(file, file.Name())
	} else {
		bufRead := bytes.NewReader(webBytes)
		name := strings.Split(filename, "/")
		newImage, depth, err = fToImg(bufRead, name[len(name)-1])
	}

	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	var extraX, extraY int
	// stretch the image if required
	if c.ImgFill != "preserve" {

		// Get wh and resize if needed if wh>xy throw an exception
		// w, h := canvas.Bounds().Max.X, canvas.Bounds().Max.Y // ImgSize()

		// if image fill is not called then move it around.

		// Replace with our own brand eventually that is true 64 bit
		// Make it 64 but it needs a proper method to change it
		w, h := resizeParams(c.ImgFill, newImage.Bounds().Max, resp.BaseImage().Bounds().Max)

		if w != newImage.Bounds().Max.X || h != newImage.Bounds().Max.Y {
			newImage = resize.Resize(uint(w), uint(h), newImage, resize.Bicubic)
			// https://pkg.go.dev/golang.org/x/image/draw#pkg-variables use a different resize
		}
	} else {
		// get centre

		extraX = (resp.BaseImage().Bounds().Dx() - newImage.Bounds().Dx()) / 2
		extraY = (resp.BaseImage().Bounds().Dy() - newImage.Bounds().Dy()) / 2
	}
	// newImg64 := gridgen.ImageGenerator(*c, image.Rect(0, 0, newImage.Bounds().Max.X, newImage.Bounds().Max.Y))
	if c.ColourSpace == nil {
		c.ColourSpace = &colour.ColorSpace{}
	}

	newImg64 := colour.NewNRGBA64(req.PatchProperties.ColourSpace, newImage.Bounds())

	imgOffset, err := c.CalcOffset(resp.BaseImage().Bounds().Max)
	imgOffset = imgOffset.Add(image.Point{X: extraX, Y: extraY})

	// imgOffset = imgOffset.Add()
	if err != nil {
		resp.Write(tsg.WidgetError, fmt.Sprintf("0DEV error extracting the image offset %v", err))
	}

	if depth == 8 {
		b := newImg64.Bounds().Max
		for x := 0; x < b.X; x++ {
			for y := 0; y < b.Y; y++ {
				got := newImage.At(x, y)

				// fullDepth := colourgen.ConvertNRGBA64(got)

				newImg64.Set(x, y, got) // fullDepth)

			}
		}
	} else {
		colour.Draw(newImg64, newImg64.Bounds(), newImage, image.Point{}, draw.Over)
	}

	// draw.Src ensures the colourspace transformations are kept
	// as long as the picture has no alpha
	colour.Draw(resp.BaseImage(), image.Rectangle{Min: resp.BaseImage().Bounds().Min.Add(imgOffset), Max: resp.BaseImage().Bounds().Max}, newImg64, image.Point{}, draw.Src)

	resp.Write(tsg.WidgetSuccess, "success")
}

func (i Config) Generate(canvas draw.Image, opts ...any) error {

	filename := i.Image
	if filename == "" {
		return fmt.Errorf("0161 No image declared")
	}

	if len(opts) != 1 {
		return fmt.Errorf("0168 incorrect number of parameters passed to add image")
	}
	c, ok := opts[0].(*context.Context)
	if !ok {
		return fmt.Errorf("0169 configuration error when assiging context toadd image")
	}
	wDir := core.GetDir(*c)
	// Just check if it's a website first
	webBytes, errOpen := core.GetWebBytes(c, filename)
	var newImage image.Image
	var err error
	var depth int
	// Open a local file next if not
	if errOpen != nil {

		file, errOpen := os.Open(filepath.Join(wDir, filename))
		if errOpen != nil {
			return fmt.Errorf("0162 %v", errOpen)
		}
		newImage, depth, err = fToImg(file, file.Name())
	} else {
		bufRead := bytes.NewReader(webBytes)
		name := strings.Split(filename, "/")
		newImage, depth, err = fToImg(bufRead, name[len(name)-1])
	}

	if err != nil {
		return err
	}

	var extraX, extraY int
	// stretch the image if required
	if i.ImgFill != "preserve" {

		// Get wh and resize if needed if wh>xy throw an exception
		// w, h := canvas.Bounds().Max.X, canvas.Bounds().Max.Y // ImgSize()

		// if image fill is not called then move it around.

		// Replace with our own brand eventually that is true 64 bit
		// Make it 64 but it needs a proper method to change it
		w, h := resizeParams(i.ImgFill, newImage.Bounds().Max, canvas.Bounds().Max)

		if w != newImage.Bounds().Max.X || h != newImage.Bounds().Max.Y {
			newImage = resize.Resize(uint(w), uint(h), newImage, resize.Bicubic)
			// https://pkg.go.dev/golang.org/x/image/draw#pkg-variables use a different resize
		}
	} else {
		// get centre

		extraX = (canvas.Bounds().Dx() - newImage.Bounds().Dx()) / 2
		extraY = (canvas.Bounds().Dy() - newImage.Bounds().Dy()) / 2
	}
	// newImg64 := gridgen.ImageGenerator(*c, image.Rect(0, 0, newImage.Bounds().Max.X, newImage.Bounds().Max.Y))
	if i.ColourSpace == nil {
		i.ColourSpace = &colour.ColorSpace{}
	}

	newImg64 := colour.NewNRGBA64(*i.ColourSpace, newImage.Bounds())

	imgOffset, err := i.CalcOffset(canvas.Bounds().Max)
	imgOffset = imgOffset.Add(image.Point{X: extraX, Y: extraY})

	// imgOffset = imgOffset.Add()
	if err != nil {
		return fmt.Errorf("0DEV error extracting the image offset %v", err)
	}

	if depth == 8 {
		b := newImg64.Bounds().Max
		for x := 0; x < b.X; x++ {
			for y := 0; y < b.Y; y++ {
				got := newImage.At(x, y)

				// fullDepth := colourgen.ConvertNRGBA64(got)

				newImg64.Set(x, y, got) // fullDepth)

			}
		}
	} else {
		colour.Draw(newImg64, newImg64.Bounds(), newImage, image.Point{}, draw.Over)
	}

	// draw.Src ensures the colourspace transformations are kept
	// as long as the picture has no alpha
	colour.Draw(canvas, image.Rectangle{Min: canvas.Bounds().Min.Add(imgOffset), Max: canvas.Bounds().Max}, newImg64, image.Point{}, draw.Src)

	return nil
}

func fToImg(file io.Reader, fname string) (img image.Image, depth int, err error) {

	regTIFF := regexp.MustCompile(`^[\w\W]{1,255}\.[tT][iI][fF]{1,2}$`)
	regPNG := regexp.MustCompile(`^[\w\W]{1,255}\.[pP][nN][gG]$`)
	// var img image.Image
	// Add checks to ensure 16 bit for png and tiffs
	switch {
	case regPNG.MatchString(fname):
		buf := &bytes.Buffer{} // Get a copy of the reader for both functions
		tee := io.TeeReader(file, buf)
		depth, err = pngFence(fname, tee)

		if err != nil {
			return
		}

		img, err = png.Decode(buf)

	case regTIFF.MatchString(fname):
		buf := &bytes.Buffer{}
		tee := io.TeeReader(file, buf)
		depth, err = tiffFence(fname, tee)
		if err != nil {
			return
		}
		img, err = tiff.Decode(buf)
	default:
		err = fmt.Errorf("0163 %s is an invalid file type", fname)

	}
	if err != nil {
		if err.Error()[:4] != "0163" {
			err = fmt.Errorf("0167 %v", err)
		}
	}

	return
}

// 16 bit checker for png and tiff

func pngFence(fname string, file io.Reader) (int, error) {

	f, _ := io.ReadAll(file)
	// Knock through 8byte header magic number
	// Byte 16 has the required information
	magicNum := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if len(f) < 25 {
		return 0, fmt.Errorf("0164 file too small")
	}
	if !reflect.DeepEqual(f[0:8], magicNum) {
		return 0, fmt.Errorf("0165 %s is an invalid PNG file", fname)
	}
	if f[24] != 16 && f[24] != 8 {
		return 0, fmt.Errorf("0166 %s colour depth is %v bits not 8/16 bits. Only 8/16 bit files are accepted", fname, f[24])
	}

	return int(f[24]), nil
}

func tiffFence(fname string, file io.Reader) (int, error) {
	// Get the file infortmation
	f, _ := io.ReadAll(file)
	var order binary.ByteOrder
	if len(f) < 24 {
		return 0, fmt.Errorf("0164 file too small")
	}
	// Establish if little endian or big endian
	switch string(f[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		// Blow the doors off if they somehow made it this far

		return 0, fmt.Errorf("0165 %s is an invalid TIFF file", fname)
	}

	// Check magic number to show it really is a tiff file
	magic := getUint16(f[2:4], order)
	if magic != 42 {
		// Blow the doors off if they somehow made it this far
		return 0, fmt.Errorf("0165 %s is an invalid TIFF file", fname)
	}

	// Find the offset of the ifd header
	ifdOff := getUint32(f[4:8], order)

	// Find how many directories there are
	ifdNum := getUint16(f[ifdOff:ifdOff+2], order)

	// Offset by 2 as the first 2 are a directory
	ifdOff += 2
	var cdepth uint16
	for i := 0; i < int(ifdNum); i++ {
		// Check the tag of the ifd
		tag := getUint16(f[ifdOff+uint32(12*i):ifdOff+uint32(12*i)+2], order)
		// Fmt.Println(tag)

		if tag == 258 {
			// Find how many
			ccount := getUint32(f[ifdOff+uint32(12*i)+4:ifdOff+uint32(12*i)+8], order)
			// Where are the colour
			cOff := getUint32(f[ifdOff+uint32(12*i)+8:ifdOff+uint32(12*i)+12], order)

			for i := 0; i < int(ccount); i++ {
				cdepth = getUint16(f[cOff+uint32(2*i):cOff+uint32(2*i)+2], order)

				if cdepth != 16 && cdepth != 8 {

					return 0, fmt.Errorf("0166 %s colour depth is %v bits not 16 bits. Only 16 bit files are accepted", fname, cdepth)
				}
			}
		}
	}

	return int(cdepth), nil
}

func getUint16(b []byte, order binary.ByteOrder) uint16 {
	var byteNum uint16
	bufMag := bytes.NewReader(b)
	_ = binary.Read(bufMag, order, &byteNum)

	return byteNum
}

func getUint32(b []byte, order binary.ByteOrder) uint32 {
	var byteNum uint32
	bufMag := bytes.NewReader(b)
	_ = binary.Read(bufMag, order, &byteNum)

	return byteNum
}

func resizeParams(resizeType string, original, target image.Point) (int, int) {
	// Break up the types here
	switch strings.ToLower(resizeType) {
	case "x scale":
		// Scale to the x
		scale := float64(target.X) / float64(original.X)
		h := math.Round(scale * float64(original.Y)) // Scale x to match

		return target.X, int(h)
	case "y scale":
		scale := float64(target.Y) / float64(original.Y)
		w := math.Round(scale * float64(original.X)) // Scale y to match

		return int(w), target.Y
	case "xy scale":
		scaleY := float64(target.Y) / float64(original.Y)
		scaleX := float64(target.X) / float64(original.X)

		var scale float64

		if scaleY > scaleX {
			scale = scaleX
		} else {
			scale = scaleY
		}

		w := math.Round(scale * float64(original.X))
		h := math.Round(scale * float64(original.Y))

		return int(w), int(h)
	default: // Case "fill":

		return target.X, target.Y
	}
}
