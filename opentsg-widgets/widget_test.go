package opentsgwidgets

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/bars"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/luma"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/nearblack"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/saturation"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/ebu3373/twosi"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/legacy"
)

func TestHandlers(t *testing.T) {

	// Run the legacy handler
	otsg, _ := tsg.BuildOpenTSG("./testdata/legacyloader.json", "", true, &tsg.RunnerConfiguration{RunnerCount: 6, ProfilerEnabled: true})
	// otsg.AddCustomWidgets(twosi.SIGenerate, nearblack.NBGenerate, bars.BarGen, saturation.SatGen, luma.Generate, zoneplate.ZoneGen)
	otsg.Handle("builtin.legacy", []byte("{}"), legacy.Legacy{})
	tsg.LogToFile(otsg, slog.HandlerOptions{Level: slog.LevelDebug}, "./testdata/", "og")
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

	tsg.LogToFile(otsgh, slog.HandlerOptions{Level: slog.LevelDebug}, "./testdata/", "handler")
	tsg.AddBaseEncoders(otsgh)
	otsgh.Run("")
}
