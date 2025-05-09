// Package tsg combines the core and widgets to draw the valeus for each frame
package tsg

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"

	ascmhl "github.com/mrmxf/opentsg-mhl"
	"github.com/mrmxf/opentsg-modules/opentsg-io/csvsave"
	"github.com/mrmxf/opentsg-modules/opentsg-io/dpx"
	"github.com/mrmxf/opentsg-modules/opentsg-io/exr"
	"github.com/mrmxf/opentsg-modules/opentsg-io/tiffup"
)

// Encoder is a function for encoding the openTSG output into a
// specified format.
type Encoder func(io.Writer, image.Image, EncodeOptions) error

// EncodeOptions contains an extra options for encoding a file
type EncodeOptions struct {
	// the target bitdepth an image is saved to
	// only relevant for DPX files
	BitDepth int
}

// HandleFunc registers the handler function for the given pattern in [DefaultServeMux].
// The documentation for [ServeMux] explains how patterns are matched.
func (o OpenTSG) EncoderFunc(extension string, encoder Encoder) {
	// set up router here
	extension = strings.ToUpper(extension)
	if _, ok := o.encoders[extension]; ok {
		panic(fmt.Sprintf("The encoder extension %s has already been declared", extension))
	}

	// do some checking for invalid characters, if there
	// are any

	o.encoders[extension] = encoder

}

/////////////////////////////
// Save function wrappers //
////////////////////////////

// writeTiffFile saves the file as a tiff
func EncodeTiffFile(w io.Writer, img image.Image, _ EncodeOptions) error {

	// check for opaque
	bound := img.Bounds()
	for x := bound.Min.X; x < bound.Max.X; x++ {
		for y := bound.Min.Y; y < bound.Max.Y; y++ {
			if _, _, _, A := img.At(x, y).RGBA(); A != 65535 {
				// if there is one bit of transparency save with this method
				return colour.TiffEncode(w, img, nil)
			}
		}
	}

	switch canvas := img.(type) {
	case *image.NRGBA64:

		return tiffup.Encode(w, canvas)
	case *colour.NRGBA64:
		return colour.TiffEncode(w, canvas.BaseImage(), nil)

	default:
		// return the alpha channel version anyway
		// as at it will save the file and not crash
		return colour.TiffEncode(w, img, nil)
	}

	// if it passes the transparency check save without
	// return tiffup.Encode(f, img.(*image.NRGBA64))

}

// EncodePngFile saves the image as a png
func EncodePngFile(w io.Writer, image image.Image, _ EncodeOptions) error {
	return colour.PngEncode(w, image)
}

func EncodeExrFile(w io.Writer, image image.Image, _ EncodeOptions) error {
	return exr.Encode(w, image)
}

func EncodeDPXFile(w io.Writer, toDraw image.Image, eo EncodeOptions) error {
	// default all files to 16 bit
	if eo.BitDepth == 0 {
		eo.BitDepth = 16
	}
	switch canvas := toDraw.(type) {
	case *image.NRGBA64:
		return dpx.Encode(w, canvas, &dpx.Options{Bitdepth: eo.BitDepth})
	case *colour.NRGBA64:
		return dpx.Encode(w, canvas.BaseImage(), &dpx.Options{Bitdepth: eo.BitDepth})
	default:
		return fmt.Errorf("configuration error image of type %v can not be saved as a dpx", reflect.TypeOf(toDraw))
	}
	// assert the image here as
	// 	return dpx.Encode(f, toDraw.(*image.NRGBA64), &dpx.Options{Bitdepth: bit})
}

func EncodeCSVFile(w io.Writer, toDraw image.Image, _ EncodeOptions) error {
	// filename := file.Name()

	switch canvas := toDraw.(type) {
	case *image.NRGBA64:
		return csvsave.Encode(w, canvas)
	case *colour.NRGBA64:
		return csvsave.Encode(w, canvas.BaseImage())
	default:
		return fmt.Errorf("configuration error image of type %v can not be saved as a csv", reflect.TypeOf(toDraw))

	}
	// return csvsave.Encode(filename, img.(*image.NRGBA64))
}

/*
Add base encoders adds the following file encoders to an OpenTSG object:

  - dpx
  - csv
  - png
  - exr
  - tiff (as tiff and tif)
*/
func AddBaseEncoders(tsg *OpenTSG) {

	tsg.EncoderFunc("dpx", EncodeDPXFile)
	tsg.EncoderFunc("csv", EncodeCSVFile)
	tsg.EncoderFunc("png", EncodePngFile)
	tsg.EncoderFunc("exr", EncodeExrFile)
	tsg.EncoderFunc("tiff", EncodeTiffFile)
	tsg.EncoderFunc("tif", EncodeTiffFile)

}

func (tsg *OpenTSG) encodeFrame(filename string, base draw.Image, opts EncodeOptions) error {

	extensions := strings.Split(filename, ".")
	ext := extensions[len(extensions)-1]

	// extract the extension type
	encodeFunc, ok := tsg.encoders[strings.ToUpper(ext)]

	if !ok {
		formats := make([]string, len(tsg.encoders))
		i := 0
		for k := range tsg.encoders {
			formats[i] = k
			i++
		}

		return fmt.Errorf("%s does not have an available encoder, available encoders are: %v", filename, formats)
	}

	// open the file if not sth or the other

	saveTarget, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return fmt.Errorf("0051 %v", err)
	}

	defer saveTarget.Close()

	// wrap the function based in a context
	var fwErr error
	encodeContext := ContFunc(func(ctx context.Context) {
		fwErr = encodeFunc(saveTarget, base, opts)

	})

	// add the middleware for the encoders
	encoder := chain(tsg.contextMiddlewares, encodeContext)

	encoder(setName(context.Background(), filename))
	if fwErr != nil {
		return fmt.Errorf("0051 %v", fwErr)
	}

	// Amend the case statement for the different types of files here.
	// This means only the open tpg code can be changed
	// and custom save functions can be plugged in.

	// get the 16 bit pixels and put it through
	canvas, ok := base.(*image.NRGBA64)
	if !ok { // set to nrgba64 if not ok
		canvas = image.NewNRGBA64(base.Bounds())
		colour.Draw(canvas, canvas.Bounds(), base, image.Point{}, draw.Src)
	}
	pixB := canvas.Pix
	// reset the file to the start for the hashreader
	_, err = saveTarget.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("0052 %v", err)
	}
	err = ascmhl.MhlGenFile(saveTarget, ascmhl.ToHash{Md5: true, C4: true, Xxh128: true, Crc32RGB: true, Crc16RGB: true}, pixB, 16)

	if err != nil {
		return fmt.Errorf("0053 %v", err)
	}
	return err
	// return saveCRC(saveTarget, pixB)

}
