package tsg

import (
	"fmt"
	"image"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBadFile(t *testing.T) {
	// testing encoding the regex for savefile of bad file formats, to ensure these are not passed and saved
	var mockCanvas *image.NRGBA

	badNames := []string{"bad.bad", "bad", "bad.pngg", "m.tiffjj", "",
		`"verylongfilenameysbLt2PmqF2g1PGzUq8PyKawd74ESN0UYZ1Vm368s20zvBBxJvPjNt1H3N2xTPSPCjKX
	1D0X9aunXqSad7IhW0Z9Bsi01Rv524J2cG8O9zhKD7F7dMcBTk7054UgcOfn8VS8D0eltbzl4TBWYmZM77yRdd
	Vg3xAC9TcJvlZyRCj916XlZyrYStXDAe6Gq0AgcpNj0WRFi83j0w9Mx7ka4InSmPvQ194y3NnAokWe68mLHh19
	QpYuOHqvC77rCfgv05QQZFnLrg2FRQB1L0E4wEuP5225qUj4Mb1ua1kZ3JGqhscXcNU6XDaoG7jPsZvhobk8Zl
	Ww1Gl.png"`}
	var badExt [7]error
	for i := range badNames {
		baseTPG := OpenTSG{customSaves: baseSaves()}
		badExt[i] = baseTPG.savefile(badNames[i], "", (*image.NRGBA64)(mockCanvas), 16)

		Convey("Checking that incorrect extensions are not sent through", t, func() {
			Convey(fmt.Sprintf("using an a name of %s ", badNames[i]), func() {
				Convey("An error is returned describing it is not a valid format", func() {
					So(badExt[i], ShouldResemble, fmt.Errorf("%s is not a valid file format, please choose one of the following: tiff, png, dpx,exr or csv", badNames[i]))
				})
			})

		})
		os.Remove(badNames[i])
	}
}

func TestGoodFile(t *testing.T) {
	// testing encoding the regex for savefile of acceptable file formats, to ensure these are not passed and saved

	mockCanvas := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{2, 2}})

	goodNames := []string{"good.pNg", "./name-with_random.tiff", "g.TIf", "這是有效的文件名.png",
		"これは有効なファイル名です.tif", "этодопустимоеимяфайла.PNG", "space file.png", "hello.dpx", "yes.png"}

	// fileD = func() int { return 16 }
	baseTPG := OpenTSG{customSaves: baseSaves()}
	for _, name := range goodNames {
		goodExt := baseTPG.savefile(name, "", (*image.NRGBA64)(mockCanvas), 16)

		Convey("Checking that correct extensions are sent through", t, func() {
			Convey(fmt.Sprintf("using an a name of %s ", name), func() {
				Convey("There is no error generated", func() {
					So(goodExt, ShouldBeNil)
				})
			})
		})
		os.Remove(name)
	}
}

func TestMustachers(t *testing.T) {
	//export LD_LIBRARY_PATH=$PWD/lib:$LD_LIBRARY_PATH

	// testing encoding the regex for savefile of acceptable file formats, to ensure these are not passed and saved

	mockCanvas := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{2, 2}})

	goodNames := []string{"secrettest{{framenumber}}.pNg", "run{{framenumber}}.png"}
	framenumbers := []string{"0000", "3452"}
	expec := []string{"./secrettest0000.pNg", "run3452.png"}
	baseTPG := OpenTSG{customSaves: baseSaves()}
	// fileD = func() int { return 16 }
	for i, name := range goodNames {

		goodExt := baseTPG.savefile(name, framenumbers[i], (*image.NRGBA64)(mockCanvas), 16)
		_, openErr := os.Open(expec[i])
		Convey("Checking that the frame number is mustached correctly", t, func() {
			Convey(fmt.Sprintf("using an a name of %s ", name), func() {
				Convey("There is no error generated as the file number us filled in", func() {
					So(goodExt, ShouldBeNil)
					So(openErr, ShouldBeNil)
				})
			})
		})
		os.Remove(expec[i])
	}
}
