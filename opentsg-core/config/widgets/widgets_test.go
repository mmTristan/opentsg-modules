package widgets

import (
	"context"
	"fmt"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	. "github.com/smartystreets/goconvey/convey"
)

var mockSchema = []byte(`{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"$id": "https://example.com/product.schema.json",
	"title": "Allow anything through for tests",
	"description": "An empty schema to allow custom structs to run through",
	"type": "object"
	}`)

type all struct { // struct of the input
	Number int            `yaml:"number"`
	Float  float64        `yaml:"float"`
	Map    map[string]any `yaml:"map"`
	Array  []any          `yaml:"array"`
}

var types = all{6, 6.01, map[string]any{"some": "map"}, []any{3, 4, "5"}}

func TestExtract(t *testing.T) {

	fc, _, _ := core.FileImport("./testdata/types_loader.json", "", true)
	c, _ := core.FrameWidgetsGenerator(fc, 0)
	// fill the struct with predicted values

	// generate the results of the struct
	frames, errs := typeWrapper(c, types, "testsuite")

	Convey("Checking that structs are filled", t, func() {
		Convey("using ./testdata/types_loader.json as the input file then extract the bytes into the all struct", func() {
			Convey(fmt.Sprintf("No error is returned and the filled struct should match %v", types), func() {
				So(errs, ShouldBeNil)
				So(frames[core.AliasIdentity{FullName: "types", ZPos: 0}], ShouldResemble, types)
			})
		})
	})
}
func TestMissed(t *testing.T) {

	fc, _, _ := core.FileImport("./testdata/types_loader.json", "", true)
	c2, _ := core.FrameWidgetsGenerator(fc, 0)
	typeWrapper(c2, types, "missed")
	missed := MissingWidgetCheck(c2)
	// actualMiss := map[core.AliasIdentity]string{{FullName: "types", ZPos: 0}: "types"}
	actualMiss := map[core.AliasIdentity]string{}
	Convey("Checking that missed structs are found", t, func() {
		Convey("using ./testdata/types_loader.json as the input file then not searching for the widget type", func() {
			Convey(fmt.Sprintf("The missed map of %v is returned", actualMiss), func() {
				So(missed, ShouldResemble, actualMiss)
			})
		})
	})
}

// type wrapper is used for passing on the type of several structs to extract widgets. So I can loop through several structs
func typeWrapper[T any](c context.Context, item T, location string) (map[core.AliasIdentity]T, []error) {
	return ExtractWidgetStructs[T](location, mockSchema, &c)
}
