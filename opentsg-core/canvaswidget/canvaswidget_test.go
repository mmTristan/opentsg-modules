package canvaswidget

import (
	"context"
	"fmt"
	"image"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	. "github.com/smartystreets/goconvey/convey"
)

type Result interface {
	int | float64
}

func TestStructExtraction(t *testing.T) {
	// run file import here to assign the global array, then check everything for an extraction
	f := config.Framesize{W: 4096, H: 2160}
	mock := ConfigVals{Name: []string{"testname.png"}, FileDepth: 16, Framesize: f, GridRows: 16,
		ImageType: "NRGBA64", BaseImage: "test.png", LineColor: "#CCDDAA", Background: "#247AE0",
		LineWidth: 23.4}
	testContext := context.Background()
	testContext = context.WithValue(testContext, generatedConfig, mock)

	funcNames := []string{"GetFileDepth", "GetCanvasType", "getFileName", "GetPictureSize",
		"GetGridRows", " GetGridColumns", "GetBaseImage", "GetFillColour", "GetLineColour",
		"GetLWidth"}
	extractedFromCont := []any{GetFileDepth(testContext), GetCanvasType(testContext), GetFileName(testContext), GetPictureSize(testContext),
		GetGridRows(testContext), GetGridColumns(testContext), GetBaseImage(testContext), GetFillColour(testContext), GetLineColour(testContext),
		GetLWidth(testContext)}
	results := []any{16, "NRGBA64", []string{"testname.png"}, image.Point{4096, 2160},
		16, 1, "test.png", "#247AE0", "#CCDDAA", 23.4}

	for i, got := range extractedFromCont {
		Convey(fmt.Sprintf("Checking that values can be extracted from the base using a function of %s", funcNames[i]), t, func() {
			Convey(fmt.Sprintf("using a mock config of %v", mock), func() {
				Convey(fmt.Sprintf("the value matches %v", results[i]), func() {
					So(got, ShouldResemble, results[i])
				})
			})
		})
	}

	schem := GetCanvasSchema()

	Convey("Checking you can extract the schema", t, func() {
		Convey("using an value of test.png in the context", func() {
			Convey("#CCDDAA is returned", func() {
				So(schem, ShouldResemble, baseschema)
			})
		})
	})

}

func TestInitStage(t *testing.T) {
	cIn, _, _ := core.FileImport("testdata/baseloader.json", "", false)
	cFrame, _ := core.FrameWidgetsGenerator(cIn, 0)
	LoopInit(&cFrame)
	f := config.Framesize{W: 4096, H: 2160}
	expected := ConfigVals{Name: []string{"testname.png"}, FileDepth: 16, Framesize: f, GridRows: 16,
		ImageType: "NRGBA64", BaseImage: "test.png", LineColor: "#CCDDAA", Background: "#247AE0", Type: "builtin.canvasoptions",
		LineWidth: 23.4}
	got := cFrame.Value(generatedConfig).(ConfigVals)

	Convey("Checking loopinit assigns a configVal", t, func() {
		Convey("run using a input of ./testdata/base.json", func() {
			Convey(fmt.Sprintf("The extracted config matches %v", expected), func() {
				So(got, ShouldResemble, expected)
			})
		})
	})

	cIn, _, _ = core.FileImport("testdata/doubleloader.json", "", false)
	cDouble, _ := core.FrameWidgetsGenerator(cIn, 0)
	err := LoopInit(&cDouble)
	expectedDoubleErr := []error{fmt.Errorf("0061 too many \"builtin.canvasoptions\" widgets have been loaded (Got 2 wanted 1), can not configure openTSG")}

	Convey("Checking loopinit registers errors", t, func() {
		Convey("run using a input of ./testdata/doubleloader.json", func() {
			Convey(fmt.Sprintf("The extracted error matches %v", expectedDoubleErr), func() {
				So(err, ShouldResemble, expectedDoubleErr)
			})
		})
	})
	/*
		run an import with one and check the expected c

		then run it to get the error
	*/

}

/*
func TestSchema(t *testing.T) {
	schemaLoader := gojsonschema.NewBytesLoader(baseschema)
	b, _ := os.ReadFile("test.json")
	documentLoader := gojsonschema.NewBytesLoader(b)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		//this stops go trying to do anything that results in a nasty crash
		fmt.Printf("0006 Invalid json input for the alias %s. The following error occurred %v", "9", err)
	} else {
		if !result.Valid() {
			e := result.Errors()
			for _, i := range e {
				fmt.Println(i.String(), "TESTTARGET")
			}
			//	fmt.Printf("%#v\n", e[0])
			fmt.Printf("0002 Invalid json input for the alias %s. The following errors occurred %v", "e", result.Errors())
		}
	}

	c := jsonschema.NewCompiler()
	err = c.AddResource("schema.json", bytes.NewReader(baseschema))

	schema, err := c.Compile("schema.json")
	tn := make(map[string]interface{})
	json.Unmarshal(b, &tn)
	err = schema.Validate(tn)

	fmt.Println(err, schema, "errs")
	//fmt.Printf("%#v\n", err)
}
*/
