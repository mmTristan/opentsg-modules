package opentsgwidgets

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/bars"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/luma"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/nearblack"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/saturation"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/twosi"
)

type Legacy struct {
	// the location of the loader
	FileLocation string `json:"fileLocation" yaml:"fileLocation"`
	// mnt is the mount point of the folder
	MNT string `json:"mnt" yaml:"mnt"`
}

func (l Legacy) Handle(resp tsg.Response, req *tsg.Request) {

	otsg, err := tsg.BuildOpenTSG(l.FileLocation, "", true, &tsg.RunnerConfiguration{RunnerCount: 6, ProfilerEnabled: true})
	otsg.AddCustomWidgets(twosi.SIGenerate, nearblack.NBGenerate, bars.BarGen, saturation.SatGen, luma.Generate)
	if err != nil {

		resp.Write(500, err.Error())
		return
	}

	// run the old program as normal
	otsg.Draw(true, l.MNT, "stdout")

	resp.Write(200, "")
}

func TestXxx(t *testing.T) {

	// Run the legacy handler
	otsg, _ := tsg.BuildOpenTSG("./testdata/legacyloader.json", "", true, &tsg.RunnerConfiguration{RunnerCount: 6, ProfilerEnabled: true})
	otsg.AddCustomWidgets(twosi.SIGenerate, nearblack.NBGenerate, bars.BarGen, saturation.SatGen, luma.Generate)
	otsg.Handle("builtin.legacy", []byte("{}"), Legacy{})
	//otsg.HandleFunc("builtin.canvasoptions", func(r1 tsg.Response, r2 *tsg.Request) { fmt.Println("ring a ding") })
	otsg.Run("")

	// run the current handler methods
	otsgh, err := tsg.BuildOpenTSG("./testdata/handlerLoader.json", "", true, &tsg.RunnerConfiguration{RunnerCount: 6, ProfilerEnabled: true})
	fmt.Println(err)
	//	otsgh.HandleFunc("builtin.canvasoptions", func(r1 tsg.Response, r2 *tsg.Request) { fmt.Println("ring a ding") })
	otsgh.Handle(bars.WidgetType, bars.Schema, bars.BarJSON{})
	otsgh.Handle(luma.WidgetType, luma.Schema, luma.LumaJSON{})
	otsgh.Handle(nearblack.WidgetType, nearblack.Schema, nearblack.Config{})
	otsgh.Handle(saturation.WidgetType, saturation.Schema, saturation.Config{})
	otsgh.Handle(twosi.WidgetType, twosi.Schema, twosi.Config{})
	jSlog := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	tsg.AddBaseEncoders(otsgh)
	otsgh.Use(tsg.Logger(slog.New(jSlog)))
	otsgh.Run("")
}
