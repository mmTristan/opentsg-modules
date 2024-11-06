// package widgets wraps core functions to be used by widgethandler
package widgets

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/validator"
	"gopkg.in/yaml.v3"
)

// widgetfactory generates a map of bytes for a certain widget type
// with the map key being their alias and zposition.
func widgetFactory(tag string, c *context.Context) map[core.AliasIdentity]json.RawMessage {
	tagBytes := make(map[core.AliasIdentity]json.RawMessage)
	frameWidgets := core.GetFrameWidgets(*c) // use  a get function here with contents exported
	for k, wf := range frameWidgets {
		if wf.Tag == tag {
			tagBytes[core.AliasIdentity{FullName: k, ZPos: wf.Pos}] = wf.Data
		}
	}

	return tagBytes
}

func ExtractAllWidgets(c *context.Context) map[core.AliasIdentity]json.RawMessage {

	frameWidgets := core.GetFrameWidgets(*c)
	tagBytes := make(map[core.AliasIdentity]json.RawMessage)
	for k, wf := range frameWidgets {

		tagBytes[core.AliasIdentity{FullName: k, ZPos: wf.Pos, WType: wf.Tag, GridAlias: wf.Alias, Location: wf.Location, ColourSpace: wf.ColourSpace}] = wf.Data

	}

	return tagBytes
}

// ExtractWidgetStructs uses generics to generate the maps for each frame,
// an error is returned for failed the validations.
// It is used by widget handler module to get the maps for each widget, and kept in config to utilise the context
func ExtractWidgetStructs[T any](ftype string, schema []byte, c *context.Context) (map[core.AliasIdentity]T, []error) {
	get := widgetFactory(ftype, c)
	base := make(map[core.AliasIdentity]T)
	var errors []error

	names := core.GetAppliedWidgets(*c) // get the names and file locations
	lineErrs := core.GetJSONLines(*c)
	names.Mu.Lock() // prevent concurrent map writes

	defer names.Mu.Unlock()

	for key, val := range get {
		k := key.FullName
		var baseWidg T

		// check it passes the schema
		err := validator.SchemaValidator(schema, val, k, lineErrs)
		if err != nil {
			errors = append(errors, err...)

		} else {
			// only assign to a widget if it passes the schema
			if err := yaml.Unmarshal(val, &baseWidg); err != nil {
				errors = append(errors, fmt.Errorf("0032 error extracting %s into Type %v : %v", k, reflect.TypeOf(baseWidg), err))
			} else {
				names.Data[k] = k
				base[key] = baseWidg
			}
		}
	}

	return base, errors
}

// MissingWidgetCheck compares all the widgets that were assigned to every widget generated for a frame, it enumerates
// a list of every missed widget and their zpos. This is so the missed ones can still be ran and not lead to blocking
// of the image generation.
func MissingWidgetCheck(c context.Context) map[core.AliasIdentity]string {
	bases := core.GetFrameWidgets(c)
	appliedWidgets := core.GetAppliedWidgets(c)
	// update name map
	missed := make(map[core.AliasIdentity]string)
	for widgetName, content := range bases {
		if appliedWidgets.Data[widgetName] == "" && content.Tag != "" { // if it wasn't assigned a tag and isn't a factory
			missed[core.AliasIdentity{FullName: widgetName, ZPos: content.Pos}] = widgetName
		}
	}

	return missed
}
