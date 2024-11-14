// Package tsg combines the core and widgets to draw the valeus for each frame
package tsg

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"

	ascmhl "github.com/mrmxf/opentsg-mhl"
	"github.com/mrmxf/opentsg-modules/opentsg-io/csvsave"
	"github.com/mrmxf/opentsg-modules/opentsg-io/dpx"
	"github.com/mrmxf/opentsg-modules/opentsg-io/exr"
	"github.com/mrmxf/opentsg-modules/opentsg-io/tiffup"
)

type Encoder func(io.Writer, image.Image, EncodeOptions) error

type EncodeOptions struct {
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

// CanvasSave saves the file according to the extensions provided
// the name add is for debug to allow to identify images
func (tpg *OpenTSG) canvasSave(canvas draw.Image, filename []string, bitdeph int, mnt, framenumber string, debug bool, logs *errhandle.Logger) {
	for _, name := range filename {
		truepath, err := filepath.Abs(filepath.Join(mnt, name))
		if err != nil {
			logs.PrintErrorMessage("E_opentsg_SAVE_", err, debug)

			continue
		}
		err = tpg.savefile(truepath, framenumber, canvas, bitdeph)
		if err != nil {
			logs.PrintErrorMessage("E_opentsg_SAVE_", err, debug)
		}
	}
}

// saveType Extensions, regex and error

func baseSaves() map[string]func(io.Writer, draw.Image, int) error {

	return map[string]func(io.Writer, draw.Image, int) error{
		"DPX": WriteDPXFile,
		"TIF": WriteTiffFile, "TIFF": WriteTiffFile,
		"PNG": WritePngFile,
		"EXR": WriteExrFile,
		"CSV": WriteCSVFile,
	}
}

/*var saveTypes = map[string]func(*os.File, draw.Image, int) error{
	"DPX": WriteDPXFile,
	"TIF": WriteTiffFile, "TIFF": WriteTiffFile,
	"PNG": WritePngFile,
	"EXR": WriteExrFile,
	"CSV": WriteCSVFile,
} */

func (tpg *OpenTSG) savefile(filename, framenumber string, base draw.Image, bitdepth int) error {
	// regTIFF := regexp.MustCompile(`^[\w\W]{1,255}\.[tT][iI][fF]{1,2}$`)
	// regPNG := regexp.MustCompile(`^[\w\W]{1,255}\.[pP][nN][gG]$`)
	// regCSV := regexp.MustCompile(`^[\w\W]{1,255}\.[cC][sS][Vv]$`)
	// regDPX := regexp.MustCompile(`^[\w\W]{1,255}\.[dD][pP][xX]$`)
	// regEXR := regexp.MustCompile(`^[\w\W]{1,255}\.[eE][xX][rR]$`)

	filename, _ = mustache.Render(filename, map[string]string{"framenumber": framenumber})

	extensions := strings.Split(filename, ".")
	ext := extensions[len(extensions)-1]

	// extract the extension type
	saveFunc, ok := tpg.customSaves[strings.ToUpper(ext)]

	if !ok {
		return fmt.Errorf("%s is not a valid file format, please choose one of the following: tiff, png, dpx,exr or csv", filename)
	}

	// open the file if not sth or the other

	saveTarget, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
	defer saveTarget.Close()
	if err != nil {
		return fmt.Errorf("0051 %v", err)
	}

	fwErr := saveFunc(saveTarget, base, bitdepth)
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

/////////////////////////////
// Save function wrappers //
////////////////////////////

// writeTiffFile saves the file as a tiff
func WriteTiffFile(w io.Writer, img draw.Image, _ int) error {

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

// writePngFile saves the file as a png
func WritePngFile(w io.Writer, image draw.Image, _ int) error {
	return colour.PngEncode(w, image)
}

// EncodePngFile saves the image as a png
func EncodePngFile(w io.Writer, image image.Image, _ EncodeOptions) error {
	return colour.PngEncode(w, image)
}

func WriteExrFile(w io.Writer, image draw.Image, _ int) error {
	return exr.Encode(w, image)
}

func EncodeExrFile(w io.Writer, image image.Image, _ EncodeOptions) error {
	return exr.Encode(w, image)
}

func WriteDPXFile(w io.Writer, toDraw draw.Image, bit int) error {
	// default all files to 16 bit
	if bit == 0 {
		bit = 16
	}
	switch canvas := toDraw.(type) {
	case *image.NRGBA64:
		return dpx.Encode(w, canvas, &dpx.Options{Bitdepth: bit})
	case *colour.NRGBA64:
		return dpx.Encode(w, canvas.BaseImage(), &dpx.Options{Bitdepth: bit})
	default:
		return fmt.Errorf("configuration error image of type %v can not be saved as a dpx", reflect.TypeOf(toDraw))
	}
	// assert the image here as
	// 	return dpx.Encode(f, toDraw.(*image.NRGBA64), &dpx.Options{Bitdepth: bit})
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

func WriteCSVFile(w io.Writer, toDraw draw.Image, _ int) error {
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

func (tsg *OpenTSG) encodeFrame(filename string, base draw.Image, bitdepth int) error {

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

	fwErr := encodeFunc(saveTarget, base, EncodeOptions{BitDepth: bitdepth})
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
