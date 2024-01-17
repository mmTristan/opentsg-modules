package examplejson

import (
	"context"
	"encoding/json"
	"image"
	"image/png"
	"os"

	"github.com/mmTristan/opentsg-modules/opentsg-core/widgethandler"
)

/*



 */

// be able to change the base some how
const (
	base = "/workspace/opentsg-modules/opentsg-widgets/exampleJson/"
)

func SaveExampleJson(example widgethandler.Generator, folder, name string, saveImage bool) {

	jsonExample, _ := json.MarshalIndent(example, "", "    ")

	// check a folder exists
	if _, err := os.Stat(base + string(os.PathSeparator) + folder); os.IsNotExist(err) {
		os.MkdirAll(base+string(os.PathSeparator)+folder, 0777)
	}

	if saveImage {
		baseImage := image.NewNRGBA64(image.Rect(0, 0, 500, 500))
		mockCont := context.Background()
		example.Generate(baseImage, &mockCont)
		fImg, _ := os.Create(base + string(os.PathSeparator) + folder + string(os.PathSeparator) + name + "-example.png")
		png.Encode(fImg, baseImage)

		// Add the type and location fields
		var updater map[string]any
		json.Unmarshal(jsonExample, &updater)
		updater["type"] = folder
		updater["grid"] = map[string]string{"location": "a1", "alias": " A demo alias"}
		jsonExample, _ = json.MarshalIndent(updater, "", "    ")

	}

	f, _ := os.Create(base + string(os.PathSeparator) + folder + string(os.PathSeparator) + name + "-example.json")
	f.Write(jsonExample)

}
