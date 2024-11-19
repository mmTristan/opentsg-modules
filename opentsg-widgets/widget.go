package opentsgwidgets

import (
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/addimage"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/bowtie"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/bars"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/luma"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/nearblack"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/saturation"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/twosi"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/fourcolour"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/framecount"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/gradients"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/noise"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/qrgen"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/resize"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/textbox"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/zoneplate"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
)

// AddBuiltinWidgets adds all builtin widgets to the openTSG engine
// It adds:
/*
 - All the ebu3373 widgets
 - Addimage
 - Bowtie
 - FourColour
 - FrameCount
 - Gradients
 - Noise
 - QR Code
 - Resize
 - TextBox
 - ZonePlate

*/
func AddBuiltinWidgets(otsg *tsg.OpenTSG) {

	// EBU3373
	otsg.Handle(bars.WidgetType, bars.Schema, bars.BarJSON{})
	otsg.Handle(luma.WidgetType, luma.Schema, luma.LumaJSON{})
	otsg.Handle(nearblack.WidgetType, nearblack.Schema, nearblack.Config{})
	otsg.Handle(saturation.WidgetType, saturation.Schema, saturation.Config{})
	otsg.Handle(twosi.WidgetType, twosi.Schema, twosi.Config{})
	// Addimage
	otsg.Handle(addimage.WidgetType, addimage.Schema, addimage.Config{})
	// Bowtie
	otsg.Handle(bowtie.WidgetType, bowtie.Schema, bowtie.Config{})
	//FourColour
	otsg.Handle(fourcolour.WidgetType, fourcolour.Schema, fourcolour.Config{})
	//FrameCount
	otsg.Handle(framecount.WidgetType, framecount.Schema, framecount.Config{})
	//Gradients
	otsg.Handle(gradients.WidgetType, gradients.Schema, gradients.Ramp{})
	//Noise
	otsg.Handle(noise.WidgetType, noise.Schema, noise.Config{})
	//QR
	otsg.Handle(qrgen.WidgetType, qrgen.Schema, qrgen.Config{})
	//Resize
	otsg.Handle(resize.WidgetType, resize.Schema, resize.Config{})
	//TextBox
	otsg.Handle(textbox.WidgetType, textbox.Schema, textbox.TextboxJSON{})
	// ZonePlate
	otsg.Handle(zoneplate.WidgetType, zoneplate.Schema, zoneplate.ZConfig{})

}
