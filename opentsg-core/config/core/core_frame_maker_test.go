package core

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v3"
)

func TestGenerateAndCreateMethods(t *testing.T) {
	inputFile := ("./testdata/frame_generate/create/sequence_create.json")
	cCreate, _, err := FileImport(inputFile, "", false)
	fmt.Println(err, inputFile)
	predictedValuesCreate := []string{"./testdata/frame_generate/create/blue_create.yaml"}

	for i, pv := range predictedValuesCreate {
		n, err := FrameWidgetsGeneratorHandle(cCreate, i)
		expec, got := genHash(n, pv)

		Convey("Checking the create arrays run and update all data within those maps", t, func() {
			Convey("using ./testdata/sequence_create.json as the input with only create updates run", func() {
				Convey(fmt.Sprintf("The file is updated to match %v with all nested strings and arrays updated using mustache", pv), func() {
					So(err, ShouldBeNil)
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}

	inputFile = "./testdata/frame_generate/generate/sequence_generate.json"
	cGen, _, _ := FileImport(inputFile, "", false)

	predictedValuesGen := []string{"./testdata/frame_generate/generate/blue_gen.yaml"}

	for i, pv := range predictedValuesGen {
		n, _ := FrameWidgetsGeneratorHandle(cGen, i)
		expec, got := genHash(n, pv)

		/*

		 */

		Convey("Checking arguments are parsed in only generate", t, func() {
			Convey("using ./testdata/sequence_generate.json as the input which only uses generate to make jsons", func() {
				Convey("The map generated from the data replicates "+pv, func() {
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}

}

func TestFactoryUpdates(t *testing.T) {
	inputFile := "./testdata/frame_generate/factory_update/sequence.json"
	c, _, _ := FileImport(inputFile, "", false)

	predictedValues := []string{"./testdata/frame_generate/factory_update/blue_factory.yaml"}

	for i, pv := range predictedValues {
		n, _ := FrameWidgetsGeneratorHandle(c, i)
		expec, got := genHash(n, pv)

		//	p, _ := os.Create(fmt.Sprintf("nameup%v.json", i))
		//	p.Write(gen)

		Convey("Checking updates parsed to a factory, update every file included in the include", t, func() {
			Convey("using ./testdata/sequnce.json as the input with an update targeting the pyramids.json factory", func() {
				Convey("Every file in pyramid in pyramiads is updated and the json file matches "+pv, func() {
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}
}

func TestSubstitutions(t *testing.T) {
	inputFile := "./testdata/frame_generate/substitution/sequence.json"
	c, _, total := FileImport(inputFile, "", false)
	fmt.Println(total)

	predictedValues := []string{"./testdata/frame_generate/substitution/result_green.yaml"}

	for i, pv := range predictedValues {
		n, _ := FrameWidgetsGeneratorHandle(c, i)

		expec, got := genHash(n, pv)

		// bar := n.Value(baseKey).(map[string]WidgetContents)

		// frameJSON := make(map[string]map[string]any)

		// for k, v := range bar {
		// 	if v.Data != nil { // fill the ones with actual data
		// 		var m map[string]any
		// 		yaml.Unmarshal(v.Data, &m)
		// 		frameJSON[k] = m
		// 	}
		// }

		// fmt.Printf("\n\n\n")
		// fmt.Println(frameJSON, "end")

		// b, _ := yaml.Marshal(frameJSON)
		// res, _ := os.Create("./testdata/frame_generate/substitution/result_green.yaml")
		// res.Write(b)

		Convey("Checking arguments are parsed as alias names", t, func() {
			Convey("using ./testdata/frame_generate/substitution/sequence.json as the input ", func() {
				Convey(fmt.Sprintf("The file is updated to match %v with the names updated with their alias", pv), func() {
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}

	inputFileErr := "./testdata/frame_generate/substitution/sequence_internal.json"
	cErr, _, _ := FileImport(inputFileErr, "", false)

	predictedValuesErr := [][]error{{fmt.Errorf("0007 missing variable \"bad mustache\" in green-{{framenumber}}-{{bad mustache}}.png at framegreen")},
		{fmt.Errorf("0007 missing variable \"swatchParams\" in green-{{framenumber}}-{{swatchParams}}.png at frameswatch.pyramid")}}

	for i, pv := range predictedValuesErr {

		_, err := FrameWidgetsGeneratorHandle(cErr, i)

		Convey("Checking mustache errors are caught and returned", t, func() {
			Convey(fmt.Sprintf("using ./testdata/frame_generate/substitution/sequence_internal.json as the input at frame %v", i), func() {
				Convey(fmt.Sprintf("Errors of %v are returned", pv), func() {
					So(err, ShouldResemble, pv)
				})
			})
		})
	}
}

func TestArraysAndDot(t *testing.T) {
	inputFiles := []string{"./testdata/frame_generate/arraysAndDot/sequence_arrays.json", "./testdata/frame_generate/arraysAndDot/sequence_arrayUp.json",
		"./testdata/frame_generate/arraysAndDot/sequence_dotpath.json", "./testdata/frame_generate/arraysAndDot/sequence_create_dotpath.json"}
	predictedValues := [][]string{{"./testdata/frame_generate/results/blue_array.yaml"}, {"./testdata/frame_generate/results/blue_arrayUp.yaml"},
		{"./testdata/frame_generate/results/blue_dotpath.yaml"}, {"./testdata/frame_generate/arraysAndDot/results_create_dotpath.yaml"}}

	explanation := []string{"The method for accessing the names array in generate {\"R\":\"[:]\"}, {\"C\":\"[:]\"}, {\"B\":\"[:1]\"}", "The methods for updating using array notation of swatch[0][0:4]",
		"dotpaths to generated json result in updates", "dotpaths to created json always run in order, the update to frame.pd is run before frame.pd.pyramid1"}

	for i, inputFile := range inputFiles {
		c, _, _ := FileImport(inputFile, "", false)

		for j, pv := range predictedValues[i] {

			n, err := FrameWidgetsGeneratorHandle(c, j)

			expec, got := genHash(n, pv)

			Convey("Checking all array and dot path arguments are handled and the correct widgets are updated", t, func() {
				Convey(explanation[i], func() {
					Convey("The file is updated to match "+pv, func() {
						So(err, ShouldBeNil)
						So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
					})
				})
			})
		}
	}

}

func TestArrayClash(t *testing.T) {
	inputFile := "./frame.json"
	madeFile := []string{"./testdata/frame_generate/arrayclash/dotpaths.json", "./testdata/frame_generate/arrayclash/arrays.json",
		"./testdata/frame_generate/arrayclash/arrays_dot.json", "./testdata/frame_generate/arrayclash/arrays_lash.json"}
	extras := []string{`, "load.swatch.pyramid1": {"testadder": "factory update"}`,
		`, "load[0][3:5]": {"testadder": "factory update"}`,
		`, "load[0][3:5]": {"testadder": "factory update"},"load.swatch.pyramid3": {"testadder": "dotpath", "dottpath":"was updated"}`,
		`,"load[0][2:5]": {"testadder": "targeted array update"} ,"load.swatch[0:3]": {"testadder": "dot path arrays"}
		,"load[0]": {"testadder": "factory array update"}`} // ,

	explanations := []string{"The dotpath in the upper factory update overwrites the dotpath called in ./testdata/frame_generate/arrayclash/frame.json",
		"The arrays in the upper factory update overwrites the arrays called in ./testdata/frame_generate/arrayclash/frame.json",
		"The arrays in the upper factory overwrite the changes made to the dot path",
		"The arrays are updated in the order of load[0], then load.swatch[0:3] then load[0][2:5]"}

	result := []string{"./testdata/frame_generate/arrayclash/dotpaths_results.yaml", "./testdata/frame_generate/arrayclash/array_results.yaml",
		"./testdata/frame_generate/arrayclash/clash_results.yaml", "./testdata/frame_generate/arrayclash/array_clash_result.yaml"}

	for i, e := range extras {

		c, err := contMocker(madeFile[i], inputFile, e)

		expec, got := genHash(c, result[i])

		// p, _ := os.Create(fmt.Sprintf("name2%v.json", i))
		// p.Write(gen)
		Convey("Checking that the hierarchy runs in the intended order of dotpath then array, dotpaths called within the factory take precedence over internal ones in child factories", t, func() {
			Convey(fmt.Sprintf("using %s as the update method", extras[i]), func() {
				Convey(explanations[i], func() {
					So(err, ShouldBeNil)
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
		os.Remove(madeFile[i])
	}

}

func TestErrors(t *testing.T) {

	badJsons := []string{",\"frame[5:7]\":{\"help\":45}", ",\"frame[3:1]\":{\"help\":45}", ",\"frame.swatch.notreal\":{\"help\":45, \"swatchType\":\"blue\"}", ",\"notreal\":{\"help\":45}"} // "./testdata/frame_generate/sequence_arrays.json", "./testdata/frame_generate/sequence_arrayUp.json", "./testdata/frame_generate/sequence_dotpath.json"}
	inputFile := "./testdata/frame_generate/error_gen.json"

	predictedValuesErr := [][]error{{fmt.Errorf("0021 no matches found for frame[5:7]")},
		{fmt.Errorf("0021 no matches found for frame[3:1]")},
		{fmt.Errorf("0018 no map values found for the dot path of frame.swatch.notreal")},
		{fmt.Errorf("0018 no map values found for the dot path of notreal")}}

	for i, bad := range badJsons {

		f, _ := os.Create(inputFile)

		mid := fmt.Sprintf(`{
			"include":[
				{"uri": "./frame.json", "name":"frame"}
			],
			"create":[
				{"frame": {"swatchType":"blue"}%s},
				{"frame": {"swatchType":"green"}}
			]
		}`, bad)

		_, _ = f.Write([]byte(mid))

		c, _, _ := FileImport(inputFile, "", false)
		_, err := FrameWidgetsGeneratorHandle(c, 0)

		Convey("Checking that errors are caught within the dotpaths", t, func() {
			Convey(fmt.Sprintf("using %s as an additional input in the json", bad), func() {
				Convey(fmt.Sprintf("The file is not updated and there is an error of %v", predictedValuesErr[i]), func() {
					So(err, ShouldResemble, predictedValuesErr[i])
				})
			})
		})
		f.Close()
	}
	os.Remove(inputFile)

	inputFiles := []string{"./testdata/frame_generate/errors/sequence_repeat.json"}
	expec := [][]error{{fmt.Errorf("0015 frame.swatch.blueR0.C0.B0 has already been generated for the parent frame.swatch")}}

	for i, inputFile := range inputFiles {
		c, _, e := FileImport(inputFile, "", false)
		fmt.Println(e)
		_, errs := FrameWidgetsGeneratorHandle(c, i)

		Convey("Checking arguments generated sequences are repeated", t, func() {
			Convey("using ./testdata/frame_generate/errors/sequence_repeat.json as the input with one repeated item in the generated section", func() {
				Convey(fmt.Sprintf("An error of %v is returned", expec[i]), func() {
					So(errs, ShouldResemble, expec[i])
				})
			})
		})
	}

	propFile := "./testdata/frame_generate2/errors/loader.json"

	testCanvas, _ := filepath.Abs("./testdata/frame_generate2/errors/canvas.json")
	testWidget, _ := filepath.Abs("./testdata/frame_generate2/errors/widget.json")
	testWidgetX, _ := filepath.Abs("./testdata/frame_generate2/errors/widget_bad_x.json")
	reg := "'^-{0,1}\\d{0,2}\\.{1}\\d{0,}%$|^-{0,1}\\d{0,2}%$|^-{0,1}(100)%$|^100\\.[0]*%$'"

	propExpec := [][]error{{fmt.Errorf("0026 Invalid type. Expected: string, given: integer at line 3 in %s, for canvas", testCanvas)},
		{fmt.Errorf("0026 Additional property grid is not allowed at line 4 in %s, for widget", testWidget)},
		{fmt.Errorf("0026 Must validate at least one schema (anyOf) at line 6 in %s, for widget_bad_x", testWidgetX),
			fmt.Errorf("0026 Does not match pattern %s at line 6 in %s, for widget_bad_x", reg, testWidgetX)}}

	for i, expec := range propExpec {
		c, _, e := FileImport(propFile, "", false)
		fmt.Println(e)
		_, errs := FrameWidgetsGeneratorHandle(c, i)

		Convey("Checking the props schema catches errors of malformed properties", t, func() {
			Convey("using ./testdata/frame_generate2/errors/loader.json as the input a sequence of malformed properties", func() {
				Convey(fmt.Sprintf("An error of %v is returned", expec), func() {
					So(errs, ShouldResemble, expec)
				})
			})
		})
	}
}

func TestCreateErrors(t *testing.T) {

	// inputFiles := []string{"./testdata/frame_generate/errors/sequence_shallow.json",
	//	"./testdata/frame_generate/errors/sequence_shallow.json"}
	expectedOutcomes := [][]error{
		{fmt.Errorf("0013 at mismatch.d.blue the number of keys 2 does not match the n dimensions of the data matrix 3")},
		{fmt.Errorf("0012 at baddata.d.blue the number of data points 3 does not match the n dimensions of the data matrix 3")}}
	inputFile := "./testdata/frame_generate/errors/sequence_shallow.json"
	for i, expected := range expectedOutcomes {

		c, _, _ := FileImport(inputFile, "", false)

		_, err := FrameWidgetsGeneratorHandle(c, i)
		//	bar := n.Value(frameHolders).(base)

		Convey("Checking that errors are caught within the dotpaths", t, func() {
			Convey(fmt.Sprintf("using %s as an additional input in the json", inputFile), func() {
				Convey(fmt.Sprintf("The file is not updated and the error is %v", expected), func() {
					So(err, ShouldResemble, expected)
				})
			})
		})

	}

}

func TestZGenerateErrors(t *testing.T) {

	generateString := `{
        "_COMMENT": "We should probably have a better syntax for mapping an N dimensional array of data to objects!",
          "name": [{"R":"[:]"}, {"CD":"[:]"}, {"B":"[:]"}],
          "action": {
           "pyramid" : {
           "d.pyramid": ["grid.location","backgroundcolor"]}
          } }`

	var mockGenerate generate
	yaml.Unmarshal([]byte(generateString), &mockGenerate)
	fmt.Print(mockGenerate)
	mockBase := base{importedFactories: map[string]factory{}, importedWidgets: make(map[string]json.RawMessage), generatedFrameWidgets: map[string]widgetContents{}}
	mockZpos := 0
	err := mockBase.factoryGenerateWidgets([]generate{mockGenerate}, "base.", nil, []int{}, &mockZpos)
	Convey("Checking that errors handling of the generate function", t, func() {
		Convey("using the bare minimum to run factory generate widgets", func() {
			Convey("An error saying no data was found is returned", func() {
				So(err, ShouldResemble, []error{fmt.Errorf("0010 no data was found for base.d")})
			})
		})
	})
	mockBase.importedWidgets["base.d"] = []byte{32}
	err = mockBase.factoryGenerateWidgets([]generate{mockGenerate}, "base.", nil, []int{}, &mockZpos)
	Convey("Checking that errors handling of the generate function", t, func() {
		Convey("using the bare minimum to run factory generate widgets", func() {
			Convey("An error saying no data was found is returned", func() {
				So(err, ShouldResemble, []error{fmt.Errorf("0011 no widgets were found for base.pyramid")})
			})
		})
	})

	// Then add data

	// try and sensibly loop through the data, have a list of bases to run through each time?
}

var mockBase = `{
    "include": [
      {
        "uri": "%s",
        "name": "load"
      }
    ],
    "create": [
      {
        "load": {}
		%s
      }
    ]
  }`

// cont mocker generates the context from the input file that is generated by the user
// the file follows the template of base
// this mocks the draw set up and allows us to run the widget
func contMocker(file, target, extra string) (context.Context, any) {
	body := fmt.Sprintf(mockBase, target, extra)
	mockFile, _ := os.Create(file)
	_, _ = mockFile.Write([]byte(body))
	mockFile.Close()

	c, _, err := FileImport(file, "", false)

	if err != nil {
		return c, err
	}
	cFrame, er := FrameWidgetsGeneratorHandle(c, 0)

	return cFrame, er
}

func TestZpos(t *testing.T) {

	inputFile := "./sequence_z.json"
	madeFile := []string{"./testdata/frame_generate/zpos/dotpaths.json", "./testdata/frame_generate/zpos/arrays.json"}
	extras := []string{`, "load.pyramid1": {"position":"not real"} , "load.pyramid2": {"position":"not real"}`,
		`,"load[0:4]": {"position":"not real"}`}
	expectedLength := 7
	expec := make(map[int]bool)

	for i := 0; i < expectedLength; i++ {
		expec[i] = true
	}
	for i, e := range extras {

		c, err := contMocker(madeFile[i], inputFile, e)

		frame := c.Value(baseKey).(map[string]WidgetContents)

		//  map the z positions
		results := make(map[int]bool)
		for _, v := range frame {
			if v.Data != nil {
				results[v.Pos] = true
			}
		}

		Convey("Checking that z positions are preserved when widgets are updated", t, func() {
			Convey(fmt.Sprintf("using %s as the update method", extras[i]), func() {
				Convey("All the z positions from 0 to 6 are updated without any further updates", func() {
					So(err, ShouldBeNil)
					So(results, ShouldResemble, expec)
				})
			})
		})
		os.Remove(madeFile[i])
	}

}

func genHash(n context.Context, pv string) (hash.Hash, hash.Hash) {
	bar := n.Value(baseKey).(map[string]WidgetContents)
	hnormal := sha256.New()
	htest := sha256.New()
	read, _ := os.ReadFile(pv)
	frameJSON := make(map[string]map[string]any)

	for k, v := range bar {
		if v.Data != nil { // fill the ones with actual data
			var m map[string]any
			yaml.Unmarshal(v.Data, &m)
			frameJSON[k] = m
		}
	}

	gen, _ := yaml.Marshal(frameJSON)

	hnormal.Write(read)
	htest.Write(append([]byte("---\n"), gen...))

	if pv == "./testdata/frame_generate/create/blue_create.yaml.0" {
		p, _ := os.Create(pv + ".yaml")
		_, _ = p.Write(gen)
	}

	return hnormal, htest
}
