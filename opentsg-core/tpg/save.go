// package framehandlers is a wrapper for the save function to only save the types allowed
// for each different tpg
package tpg

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mmTristan/opentsg-modules/opentsg-core/errHandle"

	"github.com/mmTristan/opentsg-modules/opentsg-io/csvsave"
	"github.com/mmTristan/opentsg-modules/opentsg-io/dpx"
	"github.com/mmTristan/opentsg-modules/opentsg-io/exr"
	"github.com/mmTristan/opentsg-modules/opentsg-io/tiffup"
	ascmhl "github.com/mrmxf/opentsg-mhl"
)

// CanvasSave saves the file according to the extensions provided
// the name add is for debug to allow to identify images
func (tpg *opentsg) canvasSave(canvas draw.Image, filename []string, bitdeph int, mnt, framenumber string, debug bool, logs *errhandle.Logger) {
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

func baseSaves() map[string]func(*os.File, draw.Image, int) error {

	return map[string]func(*os.File, draw.Image, int) error{
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

func (tpg *opentsg) savefile(filename, framenumber string, base draw.Image, bitdepth int) error {
	// regTIFF := regexp.MustCompile(`^[\w\W]{1,255}\.[tT][iI][fF]{1,2}$`)
	// regPNG := regexp.MustCompile(`^[\w\W]{1,255}\.[pP][nN][gG]$`)
	// regCSV := regexp.MustCompile(`^[\w\W]{1,255}\.[cC][sS][Vv]$`)
	// regDPX := regexp.MustCompile(`^[\w\W]{1,255}\.[dD][pP][xX]$`)
	// regSTH := regexp.MustCompile(`^[\w\W]{1,255}\.[7][tT][hH]$`)
	// regEXR := regexp.MustCompile(`^[\w\W]{1,255}\.[eE][xX][rR]$`)

	/*
		new layout, customary sanity check with a regexp

		get the final bit for an extension match with strings.Split. then To upper and use the map

		// do the opening here etc

		saveType := make(map[string]func(*os.File, draw.Image))

		if save, validSave := saveType[ext]; validSave {
			save(all the relevantInformation)
		}

	*/

	filename, _ = mustache.Render(filename, map[string]string{"framenumber": framenumber})

	extensions := strings.Split(filename, ".")
	ext := extensions[len(extensions)-1]

	// extract the extension type
	saveFunc, ok := tpg.customSaves[strings.ToUpper(ext)]

	if !ok {
		return fmt.Errorf("%s is not a valid file format, please choose one of the following: tiff, png, dpx,exr,7th or csv", filename)
	}

	// TODO update this to be more robust and not always done
	//canvas, ok := base.(*image.NRGBA64)
	//if !ok { //set to nrgba64 if not ok
	//	canvas = image.NewNRGBA64(base.Bounds())
	//	draw.Draw(canvas, canvas.Bounds(), base, image.Point{}, draw.Src)
	//}

	//open the file if not sth or the other
	var saveTarget *os.File
	///	if openFence(filename, regTIFF, regDPX, regPNG, regEXR) {
	var err error
	saveTarget, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return fmt.Errorf("0051 %v", err)
	}

	defer saveTarget.Close()

	fwErr := saveFunc(saveTarget, base, bitdepth)
	if fwErr != nil {
		return fmt.Errorf("0051 %v", fwErr)
	}

	// Amend the case statement for the different types of files here.
	// This means only the open tpg code can be changed
	// and custom save functions can be plugged in.
	/*
		switch {
		case regTIFF.MatchString(filename):
			fwErr = WriteTiffFile(saveTarget, canvas, 0)
		case regPNG.MatchString(filename):
			fwErr = WritePngFile(saveTarget, canvas, 0)
		case regEXR.MatchString(filename):
			fwErr = WriteExrFile(saveTarget, base, 0) //base is used as it can handle the aces
		case regDPX.MatchString(filename):
			fwErr = WriteDPXFile(saveTarget, canvas, bitdepth)
		case regSTH.MatchString(filename):
			// open the file immediatley after writing it
			// as we the c library is not compatiable with io.reader
			fwErr = WriteSthFile(saveTarget, canvas, 0)
			var err error
			saveTarget, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
			defer saveTarget.Close()
			if err != nil {
				return fmt.Errorf("0051 %v", err)
			}
		case regCSV.MatchString(filename):
			// return as we don't open the file
			err := WriteCSVFile(saveTarget, canvas, 0)
			if err != nil {
				err = fmt.Errorf("0054 %v", err)
			}
			return err
		default:
			// if the file isn't saved then don't make the crc info for it as well also let them know
			return fmt.Errorf("%s is not a valid file format, please choose one of the following: tiff, png, dpx,exr,7th or csv", filename)
		}
		if fwErr != nil {
			return fmt.Errorf("0051 %v", fwErr)
		}*/

	// get the 16 bit pixels and put it through
	canvas, ok := base.(*image.NRGBA64)
	if !ok { //set to nrgba64 if not ok
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

func openFence(file string, checkers ...*regexp.Regexp) bool {
	for _, check := range checkers {
		if check.MatchString(file) {
			return true
		}
	}
	return false
}

/////////////////////////////
// Save function wrappers //
////////////////////////////

// writeTiffFile saves the file as a tiff
func WriteTiffFile(f *os.File, img draw.Image, empty int) error {

	// check for opaque

	bound := img.Bounds()
	for x := bound.Min.X; x < bound.Max.X; x++ {
		for y := bound.Min.Y; y < bound.Max.Y; y++ {
			if _, _, _, A := img.At(x, y).RGBA(); A != 65535 {
				// if there is one bit of transparency save with this method
				return colour.TiffEncode(f, img, nil)
			}
		}
	}

	switch canvas := img.(type) {
	case *image.NRGBA64:

		return tiffup.Encode(f, canvas)
	case *colour.NRGBA64:
		return colour.TiffEncode(f, canvas.BaseImage(), nil)

	default:
		// return the alpha channel version anyway
		//as at it will save the file and not crash
		return colour.TiffEncode(f, img, nil)
	}

	// if it passes the transparency check save without
	//return tiffup.Encode(f, img.(*image.NRGBA64))

}

// writePngFile saves the file as a png
func WritePngFile(f *os.File, image draw.Image, empty int) error {
	return colour.PngEncode(f, image)
}

func WriteExrFile(f *os.File, image draw.Image, empty int) error {
	return exr.Encode(f, image)
}

func WriteDPXFile(f *os.File, toDraw draw.Image, bit int) error {
	// default all files to 16 bit
	if bit == 0 {
		bit = 16
	}
	switch canvas := toDraw.(type) {
	case *image.NRGBA64:
		return dpx.Encode(f, canvas, &dpx.Options{Bitdepth: bit})
	case *colour.NRGBA64:
		return dpx.Encode(f, canvas.BaseImage(), &dpx.Options{Bitdepth: bit})
	default:
		return fmt.Errorf("configuration error image of type %v can not be saved as a dpx", reflect.TypeOf(toDraw))
	}
	//assert the image here as
	// 	return dpx.Encode(f, toDraw.(*image.NRGBA64), &dpx.Options{Bitdepth: bit})
}

func WriteCSVFile(file *os.File, toDraw draw.Image, empty int) error {
	filename := file.Name()

	switch canvas := toDraw.(type) {
	case *image.NRGBA64:
		return csvsave.Encode(filename, canvas)
	case *colour.NRGBA64:
		return csvsave.Encode(filename, canvas.BaseImage())
	default:
		return fmt.Errorf("configuration error image of type %v can not be saved as a csv", reflect.TypeOf(toDraw))

	}
	// return csvsave.Encode(filename, img.(*image.NRGBA64))
}
