// package validator is used for validating input and extracting file locations and line numbers for errors.
package validator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/internal/get"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type fileAndLocation struct {
	file, position, debug string
}

// JsonLine is the map used to store the file locations and line number for each json object value.
// The xxh64 hash of the bytes of each value on a line are used as the keys.
// Multiple identical values would then be stored with the same hash and have an option of
// file locations.
type JSONLines map[uint64]fileAndLocation

// Liner extracts all the keys from a json file with the line number(s) at which they occur and updates positions
// with their values. Each file should be written into the same JSONLines object
func Liner(file []byte, fileName, assignType string, positions JSONLines) error {

	// yaml node contains the line number for each node
	var node yaml.Node
	err := yaml.Unmarshal(file, &node)

	switch {
	case err != nil:

		return fmt.Errorf("0028 %v for extracting the yaml bytes from %s", err, fileName) // do nothing basically
	case assignType == "factory":
		err = factoryLines(fileName, node, positions)
		if err != nil {

			return err
		}
	default: // else loop through the content as only factory has the special runtime rules
		for _, cont := range node.Content {
			err = yamlLines("", fileName, *cont, positions)
			if err != nil {

				return err
			}
		}
	}

	return nil
}

// factoryLines loops through a factory object treating like yamlLines until the create map is reached.
// then the create values are passed through without a parent so they can be picked up by other widgets in the schema validator.
func factoryLines(fileName string, node yaml.Node, positions JSONLines) error {
	if len(node.Content) != 0 {
		fact := node.Content[0]
		for i := 0; i < len(fact.Content); i += 2 {
			val := fact.Content[i]

			if val.Value == "create" {
				// format it all
				// check each map adding that to pos
				creator := fact.Content[i+1]
				// loop through the creator map as a function
				for i := 0; i < len(creator.Content); i++ {
					target := creator.Content[i]
					for j := 0; j < len(target.Content); j += 2 {
						name := "create." + target.Content[j].Value
						val := target.Content[j+1]

						valBytes, err := yamlCleanse[map[string]any](val, fileName)
						if err != nil {
							return err
						}

						// get each create hash value, then send the children through with a clean slate
						xxh64 := xxhash.Sum64(append([]byte(name), valBytes...))
						posAdder(fileName, name, val.Line, xxh64, positions)
						err = yamlLines("", fileName, *target.Content[j+1], positions)
						if err != nil {

							return err
						}
					}
				}
			} else if len(fact.Content[i+1].Content) != 0 {
				// loop run the content values of the other bits
				err := yamlLines(val.Value, fileName, *fact.Content[i+1].Content[0], positions)
				if err != nil {

					return err
				}
			}

		}
	}

	return nil
}

// yamlLines recurisvely loops through all of the key values of the yaml node, assigning the line numebr for the key values.
// It generates the dot paths for the keys as well going depth first.
func yamlLines(parent, file string, child yaml.Node, positions JSONLines) error {

	if len(child.Content)%2 != 0 {

		return fmt.Errorf("invalid number of yaml children in %s got %v when expecting an even number", file, len(child.Content))
	}
	for i := 0; i < len(child.Content); i += 2 {
		tag := child.Content[i]
		name := parent + tag.Value

		value := child.Content[i+1]
		// extract the value as bytes then update if required
		valBytes, err := yaml.Marshal(value)
		if err != nil {

			return fmt.Errorf("0029 error ensuring the file information is stored as yaml %v at %s", err, file)
		}

		switch value.Tag {
		case "!!map":
			err = yamlLines(name+".", file, *value, positions)
			if err != nil {

				return err
			}
			// convert to bytes and back to bytes to unpredictable json formatting
			valBytes, err = yamlCleanse[map[string]any](value, file)
			if err != nil {

				return err
			}

		case "!!seq":
			// convert to a basic yaml style
			var values []any
			for i, arrVal := range value.Content {
				// inlcude the array position as part of the dot path
				namePos := fmt.Sprintf("%s.%v", name, i)

				if arrVal.Tag == "!!map" {
					// parse it along as an array adding the position to match
					err = yamlLines(namePos+".", file, *arrVal, positions)
					if err != nil {
						return err
					}

					// parse an empty one in as well just to collect things that may be used as updates e.g. in data jsons
					err = yamlLines("", file, *arrVal, positions)
					if err != nil {
						return err
					}
				} // add the value to the array
				values = append(values, arrVal)

				valBytes, err = yamlCleanse[any](arrVal, file)
				if err != nil {

					return err
				}

				xxh64 := xxhash.Sum64(append([]byte(namePos), valBytes...))
				posAdder(file, namePos, arrVal.Line, xxh64, positions)
			}
			// generate the array value
			valBytes, err = yamlCleanse[[]any](values, file)
			if err != nil {

				return err
			}

		case "!!str":
			if valBytes[0] == byte(rune('"')) { // check if it has the json form
				valBytes = append(valBytes[1:len(valBytes)-2], []byte{10}...) // trim quotation marks and add the endline.
			}
		} // else yaml encode the children for the name
		xxh64 := xxhash.Sum64(append([]byte(name), valBytes...))
		posAdder(file, name, tag.Line, xxh64, positions)
	}

	return nil
}

// yamlCleanse takes a yaml of any type and marshals then unmarshalls it to remove
// any of the json formatting for later comparisons.
func yamlCleanse[T any](v any, fileName string) ([]byte, error) {

	anyStyle, err := yaml.Marshal(v)
	if err != nil {

		return nil, fmt.Errorf("0029 error ensuring the file information is stored as yaml %v at %s", err, fileName)
	}

	var mid T
	err = yaml.Unmarshal(anyStyle, &mid)
	if err != nil {

		return nil, fmt.Errorf("0029 error ensuring the file information is stored as yaml %v at %s", err, fileName)
	}

	clean, err := yaml.Marshal(mid)
	if err != nil {

		return nil, fmt.Errorf("0029 error ensuring the file information is stored as yaml %v at %s", err, fileName)
	}

	return clean, nil
}

// posAdder adds the file and line to the hash map
func posAdder(file, header string, line int, xxh64 uint64, positions JSONLines) {
	if prev, ok := positions[xxh64]; ok { // if similar return chains of where this is found for ease of debugging

		var match bool

		// split regardless as an array is still formed
		files, lines := strings.Split(prev.file, ","), strings.Split(prev.position, ",")
		for i := range files {
			if files[i] == file && lines[i] == fmt.Sprintf("%v", line) {
				match = true

				break
			}
		}

		if !match { // if it has not already been added, add it
			positions[xxh64] = fileAndLocation{fmt.Sprintf("%s,%s", prev.file, file), fmt.Sprintf("%s,%v", prev.position, line), header}
		}

	} else {
		positions[xxh64] = fileAndLocation{file, fmt.Sprintf("%v", line), header}
	}
}

// schemaValidator validates a json input against the relevant schema. It returns the line number of any errors using the
// JsonLines Map.
func SchemaValidator(schema, input []byte, inputID string, fileLines JSONLines) []error {
	schemaLoader := gojsonschema.NewBytesLoader(schema)
	// cleanse input to json if initially yaml
	if !json.Valid(input) { // if not json open it as yaml and save as json
		var clean any
		err := yaml.Unmarshal(input, &clean)
		if err != nil {

			return []error{fmt.Errorf("0030 extracting %s: %v", inputID, err)}
		}

		input, err = json.Marshal(clean)
		if err != nil {

			return []error{fmt.Errorf("0031 cannot convert yaml to json at %s: %v", inputID, err)}
		}
	}

	documentLoader := gojsonschema.NewBytesLoader(input)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {

		// this stops go trying to do anything that results in a nasty crash
		return []error{fmt.Errorf("0025 Invalid json input for the alias %s. The following error occurred %v", inputID, err)}
	} else if !result.Valid() {

		return errorExtractor(result.Errors(), input, inputID, fileLines)
	}

	return nil
}

// errorExtractor wraps all the errors from json schema with their yaml/json line number and file.
func errorExtractor(schemaErrs []gojsonschema.ResultError, input []byte, loc string, fileLines JSONLines) []error {
	errs := make([]error, len(schemaErrs))

	for i, schemaErr := range schemaErrs {
		fault := schemaErr.Details()["property"]
		if fault == nil {
			fault = schemaErr.Details()["field"]
		} else if schemaErr.Details()["field"] != "(root)" { // root is just an empty space in dot form
			// join the field to its property to get the full path
			fault = schemaErr.Details()["field"].(string) + "." + fault.(string) // update for complete path
		}

		name := fault.(string)

		errorTarget, err := getJSONLines(fileLines, schemaErr.Value(), name, input)
		switch {
		case errorTarget == fileAndLocation{}: // if nothing is found then check the field
			name := schemaErr.Details()["field"].(string) // to see if a matching hash can be extracted
			errorTarget, err = getJSONLines(fileLines, schemaErr.Value(), name, input)
			if err != nil {
				errs[i] = fmt.Errorf("0033 encountered %v when looking for cause of %v in %s", err, schemaErr.Description(), loc)

				continue
			}
		case err != nil:
			errs[i] = fmt.Errorf("0033 encountered %v when looking for cause of %v in %s", err, schemaErr.Description(), loc)

			continue
		}

		if (errorTarget == fileAndLocation{}) {
			// last ditch of extract anything that matches the offending value
			var m map[string]any
			errs[i] = fmt.Errorf("0027 %v in unknown files please check your files for the %s property in the name %s", schemaErr.Description(), name, loc)
			if err := yaml.Unmarshal(input, &m); err != nil {

				continue // move onto the next layer it can't be saved as a map
			}
			// get the badtarget
			badm, _ := getTarget(m, strings.Split(name, "."))
			// save as unknown as it may be updated

			// if it's a map then find the values and hope you can get something out of it that matches
			if bm, ok := badm.(map[string]any); ok {
				// get all the possible outcomes
				allVals, allKeys := get.Get(bm, []string{}, true)
				for j, allV := range allVals {
					// get the result
					res, _ := yaml.Marshal(allV)
					// update the full path and calculate hash
					newName := name + "." + strings.Join(allKeys[j], ".")
					problemKey := xxhash.Sum64(append([]byte(newName), res...))
					errorTarget = fileLines[problemKey] // if it matches update the error and move on
					if (errorTarget != fileAndLocation{}) {
						// update the error message giving some idea of the problem then quit
						errs[i] = fmt.Errorf("0026 %v at line %v in %s, for %s", schemaErr.Description(), errorTarget.position, errorTarget.file, loc)

						break
					}
				}
			}
		} else {
			errs[i] = fmt.Errorf("0026 %v at line %v in %s, for %s", schemaErr.Description(), errorTarget.position, errorTarget.file, loc)
		}
	}

	return errs
}

// get jsonlines wraps the map valueExtract process and calculating the hash
func getJSONLines(fileLines JSONLines, value any, dotPathName string, input []byte) (fileAndLocation, error) {
	target, name, err := mapValsExtract(value, dotPathName, input)
	if err != nil {

		return fileAndLocation{}, err
	}

	problemKey := xxhash.Sum64(append([]byte(name), target...))

	return fileLines[problemKey], nil
}

// mapvals returns the yaml bytes of a map value and the full name to the key
func mapValsExtract(initVal any, path string, input []byte) ([]byte, string, error) {
	var res []byte
	// process the value depending what it is to maatch the
	switch vals := initVal.(type) {
	case string:
		var err error
		res, err = yaml.Marshal(vals)
		if err != nil {

			return res, path, err
		}

	default: // else manually search for the value
		inputAsMap := make(map[string]any)
		err := yaml.Unmarshal(input, inputAsMap)
		if err != nil {

			return res, path, err
		}

		val, runoff := getTarget(inputAsMap, strings.Split(path, "."))
		if _, ok := val.([]interface{}); ok { // name sorter outer
			names := strings.Split(path, ".")
			path = strings.Join(names[:len(names)-runoff], ".")
		}
		res, _ = yaml.Marshal(val)
	}

	return res, path, nil
}

// get target returns the value of a map and how many keys the result was overshot by.
// the overshot value is in because of the presence in arrays
func getTarget(m map[string]any, prevKeys []string) (any, int) {
	if len(prevKeys) == 0 {
		return m, 0
	}

	next := m[prevKeys[0]]
	if len(prevKeys) == 1 {
		return next, 0
	}
	switch child := next.(type) {
	case map[string]any:
		return getTarget(child, prevKeys[1:])
	case []any:
		prevKeys = prevKeys[1:]
		if len(prevKeys) == 0 {
			return next, 0
		}
		intVar, _ := strconv.Atoi(prevKeys[0])
		if intVar >= len(child) { // if it's out of range just give back the array
			return next, len(prevKeys)
		}
		arrayVal := child[intVar]

		if array, ok := arrayVal.(map[string]any); ok {
			return getTarget(array, prevKeys[1:])
		}

		return arrayVal, len(prevKeys)

	default:

		return next, len(prevKeys[1:])
	}
}
