package core

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/credentials"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v3"
)

func init() {
	location, _ = os.Getwd()
	sep = string(os.PathSeparator)
}

var location string
var sep string

// make a test for the json init stage
func TestFileRead(t *testing.T) {

	_, frames, realfile := FileImport("./testdata/newfactory.json", "", false)
	// z := GetBlackVal(0)
	Convey("Checking an exisiting file is read", t, func() {
		Convey("using ./testdata/newfactory.json as the input file", func() {
			Convey("No error is returned and the number of frames returned should be 2", func() {
				So(realfile, ShouldBeNil)
				So(frames, ShouldEqual, 2)
			})
		})
	})

	badFile := []string{"./testdata/fake.json", "./testdata/apitest.png", "", "./testdata/repeatalias.json", "./testdata/frame_generate/errors/sequence_recurse.json"}
	badFileErr := []string{fmt.Sprintf("0001 open %s%stestdata%sfake.json: no such file or directory", location, sep, sep),
		fmt.Sprintf("0028 yaml: invalid leading UTF-8 octet for extracting the yaml bytes from %s%stestdata%sapitest.png", location, sep, sep),
		fmt.Sprintf("0001 read %s: is a directory", location), "0006 the alias robocorner is repeated, every alias is required to be unique",
		"0004 recursive set initialisation file detected, the maximum dotpath depth of 30 has been reached",
	}

	for i := range badFile {
		fmt.Println(location)
		_, _, err := FileImport(badFile[i], "", false)
		Convey("Checking if bad files are read and the errors are returned", t, func() {
			Convey(fmt.Sprintf("using %v as the input file", badFile[i]), func() {
				Convey(fmt.Sprintf("An error of %v is returned", badFileErr[i]), func() {
					So(err.Error(), ShouldEqual, badFileErr[i])
				})
			})
		})
	}
}

func TestBadJson(t *testing.T) {
	testFolderLocation := fmt.Sprintf("%stestdata%swrong%s", sep, sep, sep)

	badFiles := []string{"./testdata/wrong/apiinval.json", "./testdata/wrong/empty.json",
		"./testdata/wrong/badinclude.json", "./testdata/wrong/badincludebase.json"}
	results := []string{fmt.Sprintf("0003 No frames declared in %s%sapiinval.json", location, testFolderLocation),
		fmt.Sprintf("0002 yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into core.factory when opening %s%sempty.json", location, testFolderLocation),
		fmt.Sprintf("0003 No frames declared in %s%sbadinclude.json", location, testFolderLocation),
		fmt.Sprintf("0002 yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into core.factory when opening %s%sbadincludebase.json", location, testFolderLocation)}

	for i, bf := range badFiles {
		_, _, realfile := FileImport(bf, "", false)

		Convey("Checking errors are caught for invalid json", t, func() {
			Convey(fmt.Sprintf("using %v as the input file", bf), func() {
				Convey(fmt.Sprintf("%v is returned", results[i]), func() {
					So(realfile.Error(), ShouldResemble, results[i])
				})
			})
		})
	}
}

func TestJsonRread(t *testing.T) {
	inputFile := "./testdata/frame_generate/sequence.json"
	c, _, _ := FileImport(inputFile, "", false)

	predictedValues := []string{"./testdata/frame_generate/results/blue.yaml", "./testdata/frame_generate/results/green.yaml"}

	for i, pv := range predictedValues {
		n, _ := FrameWidgetsGenerator(c, i)

		expec, got := genHash(n, pv)

		Convey("Checking arguments are parsed correctly both in the create and generate section of json factories", t, func() {
			Convey(fmt.Sprintf("Using frame %v ./testdata/sequnce.json as the input ", i), func() {
				Convey("The generated widget map as a json body matches "+pv, func() {
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}
}
func TestYamlRead(t *testing.T) {
	inputYamls := []string{"./testdata/frame_generate/yaml_test/sequence.yaml",
		"./testdata/frame_generate/yaml_test/sequence_frame.yaml", "./testdata/frame_generate/yaml_test/sequence_full.yaml"}

	predictedValuesYaml := []string{"./testdata/frame_generate/results/blue.yaml", "./testdata/frame_generate/results/green.yaml"}
	yamlMix := []string{"an input yaml file", "a mix of yaml and json files", "a complete set of yaml files"}
	for j, inputYaml := range inputYamls {
		cYaml, _, _ := FileImport(inputYaml, "", false)

		for i, pv := range predictedValuesYaml {
			n, _ := FrameWidgetsGenerator(cYaml, i)

			expec, got := genHash(n, pv)

			Convey("Checking arguments are parsed correctly both in the create and generate section of yaml json factories", t, func() {
				Convey(fmt.Sprintf("using frame %v of %s this contains a %s", i, inputYaml, yamlMix[j]), func() {
					Convey("The generated widget map as a json body matches "+pv, func() {
						So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
					})
				})
			})
		}
	}

	newDesign := "./testdata/frame_generate2/sequence.json"
	cYaml, _, e := FileImport(newDesign, "", false)
	fmt.Println(e, "input error")
	predictedValues := []string{"./testdata/frame_generate/results/blue.yaml", "./testdata/frame_generate/results/green.yaml"}

	for i, pv := range predictedValues {
		n, es := FrameWidgetsGenerator(cYaml, i)
		fmt.Println(es, "second erro")
		expec, got := genHash(n, pv)
		bar := n.Value(baseKey).(map[string]widgetContents)

		frameJSON := make(map[string]map[string]any)

		for k, v := range bar {
			if v.Data != nil { // fill the ones with actual data
				var m map[string]any
				yaml.Unmarshal(v.Data, &m)
				frameJSON[k] = m
			}
		}

		fmt.Printf("\n\n\n")
		fmt.Println(frameJSON, "end")

		b, _ := json.Marshal(frameJSON)
		res, _ := os.Create("./testdata/frame_generate2/res.json")
		res.Write(b)

		Convey("Checking arguments are parsed correctly both in the create and generate section of json factories with the new method", t, func() {
			Convey(fmt.Sprintf("Using frame %v ./testdata/sequnce.json as the input ", i), func() {
				Convey("The generated widget map as a json body matches "+pv, func() {
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}

	newDesignRoot := "./testdata/frame_generate2/sequenceRootMustache.json"
	cYamlRoot, _, e := FileImport(newDesignRoot, "", false)
	predictedValuesRoot := []string{"./testdata/frame_generate2/results/resRoot.yaml", "./testdata/frame_generate2/results/resRootNoMustache.yaml"}

	for i, pv := range predictedValuesRoot {
		n, es := FrameWidgetsGenerator(cYamlRoot, i)

		expec, got := genHash(n, pv)
		bar := n.Value(baseKey).(map[string]widgetContents)

		frameJSON := make(map[string]map[string]any)

		for k, v := range bar {
			if v.Data != nil { // fill the ones with actual data
				var m map[string]any
				yaml.Unmarshal(v.Data, &m)
				frameJSON[k] = m
			}
		}

		fmt.Printf("\n\n\n")
		fmt.Println(frameJSON, "end")

		//	b, _ := yaml.Marshal(frameJSON)
		//	res, _ := os.Create("./testdata/frame_generate2/resRoot.yaml")
		//	res.Write(b)

		Convey("Checking arguments are parsed correctly into base widgets, so widgets with declared args are updated", t, func() {
			Convey(fmt.Sprintf("Using frame %v ./testdata/frame_generate2/sequenceRootMustache.json as the input ", i), func() {
				Convey("The generated widget map as a json body matches "+pv, func() {
					So(e, ShouldBeNil)
					So(es, ShouldBeNil)
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}
}

func TestSearchOrder(t *testing.T) {

	expectedResult := []string{"./testdata/searchpath/first/first.json", "./testdata/searchpath/second/second.json", "./testdata/searchpath/third/third.json"}

	mockCredentials, _ := credentials.AuthInit("")
	parentFolders := []string{"./testdata/searchpath/", "./testdata/searchpath/first/", "./testdata/searchpath/second/", "./testdata/searchpath/third/"}
	URIList := []string{"first.json", "second.json", "third.json"}

	for i, URI := range URIList {

		expectedRes, _ := os.ReadFile(expectedResult[i])
		res, _, err := FileSearch(mockCredentials, URI, "", parentFolders)

		Convey("Checking the search order runs in the correct order", t, func() {
			Convey(fmt.Sprintf("Looking for %s, this should be found in %s", URI, expectedResult[i]), func() {
				Convey("The file is found and no error is returned ", func() {
					So(err, ShouldBeNil)
					So(res, ShouldResemble, expectedRes)
				})
			})
		})
	}

	expectedResWrkDir, _ := os.ReadFile("./testdata/searchpath/wrkdir.json")
	resWrkDir, _, errWrkDir := FileSearch(mockCredentials, "./testdata/searchpath/wrkdir.json", "", parentFolders)

	Convey("Checking the search order checks relative to the location of the executable", t, func() {
		Convey("Looking for ./testdata/searchpath/wrkdir.json", func() {
			Convey("The file is found and no error is returned ", func() {
				So(errWrkDir, ShouldBeNil)
				So(resWrkDir, ShouldResemble, expectedResWrkDir)
			})
		})
	})

	// ENVIRONMENT TEST

	holder := os.Getenv("OPENTSG_HOME")
	os.Setenv("OPENTSG_HOME", "./testdata/searchpath/env")
	expectedResEnv, _ := os.ReadFile("./testdata/searchpath/env/env.json")
	resEnv, _, errEnv := FileSearch(mockCredentials, "./env.json", "", parentFolders)

	Convey("Checking the search order checks relative to ENV value", t, func() {
		Convey("Looking for env.json relative to the environment of ./testdata/searchpath/env ", func() {
			Convey("The file is found and no error is returned ", func() {
				So(errEnv, ShouldBeNil)
				So(resEnv, ShouldResemble, expectedResEnv)
			})
		})
	})

	os.Setenv("OPENTSG_HOME", holder)
}

func TestMetadataUpdate(t *testing.T) {

	newDesignRoot := "./testdata/frame_generate2/metadataUpdates/sequence.json"
	cYamlRoot, _, _ := FileImport(newDesignRoot, "", false)
	predictedValuesRoot := []string{"./testdata/frame_generate2/metadataUpdates/resRoot.yaml"}

	for i, pv := range predictedValuesRoot {
		n, _ := FrameWidgetsGenerator(cYamlRoot, i)

		expec, got := genHash(n, pv)
		bar := n.Value(baseKey).(map[string]widgetContents)

		frameJSON := make(map[string]map[string]any)

		for k, v := range bar {
			if v.Data != nil { // fill the ones with actual data
				var m map[string]any
				yaml.Unmarshal(v.Data, &m)
				frameJSON[k] = m
			}
		}

		fmt.Printf("\n\n\n")
		fmt.Println(frameJSON, "end")

		// b, _ := yaml.Marshal(frameJSON)
		// res, _ := os.Create("./testdata/frame_generate2/metadataUpdates/resRoot2.yaml")
		// res.Write(b)

		Convey("Checking arguments are mustached with previous, so arguments can be built upon", t, func() {
			Convey(fmt.Sprintf("Using frame %v ./testdata/frame_generate2/metadataUpdates/sequence.json as the input ", i), func() {
				Convey("The generated widget map as a json body matches "+pv, func() {
					So(expec.Sum(nil), ShouldResemble, got.Sum(nil))
				})
			})
		})
	}

	newDesignError := "./testdata/frame_generate2/metadataUpdates/sequenceErr.json"
	cYamlRootErr, _, _ := FileImport(newDesignError, "", false)
	predictedErrors := []string{"0007 missing variable \"title\" in TestTitle-{{title}} at frame.canvas",
		"0007 missing variable \"update\" in {{update}}-{{title}} at frame.canvas"}

	for i, pv := range predictedValuesRoot {
		_, es := FrameWidgetsGenerator(cYamlRootErr, i)

		Convey("Checking arguments are mustached with previous, so arguments can be built upon", t, func() {
			Convey(fmt.Sprintf("Using frame %v ./testdata/frame_generate2/metadataUpdates/sequence.json as the input ", i), func() {
				Convey("The generated widget map as a json body matches "+pv, func() {
					So(es, ShouldResemble, []error{fmt.Errorf(predictedErrors[i])})
				})
			})
		})
	}
}
