package legacy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	. "github.com/smartystreets/goconvey/convey"
)

func TestXxx(t *testing.T) {

	fileLocations := []string{
		"./testdata/legacyLoaders/loadergrid.json", "./testdata/verdeLoaders/loader.json",
	}

	out := []string{"ebuLegacy.png", "myfirstTSG.png"}

	for i, location := range fileLocations {

		stdResp := tsg.TestResponder{}
		mntResp := tsg.TestResponder{}
		Legacy{FileLocation: location}.Handle(&stdResp, nil)
		Legacy{FileLocation: location, MNT: "./testdata/"}.Handle(&mntResp, nil)

		std, sErr := os.ReadFile("./testdata/" + out[i])
		mnt, mErr := os.ReadFile("./testdata/testdata/" + out[i])

		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(std)
		htest.Write(mnt)

		Convey("Checking the legacy widget run an generate an image", t, func() {
			Convey(fmt.Sprintf("Run using a loader of %s comparing the output to a mnt version", location), func() {
				Convey("The 2 versions are identical and generate the image", func() {
					So(sErr, ShouldBeNil)
					So(mErr, ShouldBeNil)
					So(stdResp.Status, ShouldResemble, tsg.WidgetSuccess)
					So(mntResp.Status, ShouldResemble, tsg.WidgetSuccess)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

		os.Remove("./testdata/" + out[i])
		os.Remove("./testdata/testdata/" + out[i])
	}

}
