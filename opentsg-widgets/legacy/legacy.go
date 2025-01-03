package legacy

import (
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/addimage"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/bowtie"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/bars"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/luma"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/nearblack"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/saturation"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/twosi"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/fourcolour"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/framecount"
	geometrytext "github.com/mrmxf/opentsg-modules/opentsg-widgets/geometryText"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/gradients"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/noise"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/qrgen"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/resize"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/textbox"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/zoneplate"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
)

// Legacy contains the simple input profile to be used
// for running the old version of openTSG.
type Legacy struct {
	// the location of the loader to be loaded into openTSG
	FileLocation string `json:"fileLocation" yaml:"fileLocation"`
	// mnt is the mount point of the folder
	MNT string `json:"mnt" yaml:"mnt"`
}

const WidgetType = "builtin.legacy"

// Handle runs the legacy version of openTSG as an independent widget
func (l Legacy) Handle(resp tsg.Response, req *tsg.Request) {

	otsg, err := tsg.BuildOpenTSG(l.FileLocation, "", true, &tsg.RunnerConfiguration{RunnerCount: 6, ProfilerEnabled: true})
	if err != nil {

		resp.Write(500, err.Error())
		return
	}

	otsg.AddCustomWidgets(twosi.SIGenerate, nearblack.NBGenerate, bars.BarGen, saturation.SatGen,
		luma.Generate, textbox.TBGenerate,
		gradients.RampGenerate, noise.NGenerator, widgethandler.MockCanvasGen,
		addimage.ImageGen, zoneplate.ZoneGen,
		framecount.CountGen, qrgen.QrGen,
		fourcolour.FourColourGenerator, geometrytext.LabelGenerator,
		bowtie.SwirlGen, resize.Gen,
		// This one should be placed last as it is checking for missed names,
		// however order doesn't matter for concurrent functions with the wait groups.
		widgethandler.MockMissedGen)

	// run the old program as normal
	otsg.Draw(true, l.MNT, "stdout")

	resp.Write(tsg.WidgetSuccess, "")
}
