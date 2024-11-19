package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config/validator"
	"github.com/mrmxf/opentsg-modules/opentsg-core/credentials"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"gopkg.in/yaml.v3"
)

// File import reads a factory json file and extracts all the included files and factories recursively.
// It returns a context holding them, the number of frames to be run and any errors encountered.
func FileImport(inputFile, profile string, debug bool, httpKeys ...string) (context.Context, int, error) {
	cont := context.Background()
	authDecoder, err := credentials.AuthInit(profile, httpKeys...)
	if err != nil {
		return cont, 0, fmt.Errorf("0000 %v", err)
	}
	inputFile, _ = filepath.Abs(inputFile)
	inputBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return cont, 0, fmt.Errorf("0001 %v", err)
	}

	data := make(validator.JSONLines)
	err = validator.Liner(inputBytes, inputFile, "factory", data)
	if err != nil { // return just the first error? figure out error handling
		return cont, 0, err
	}

	errs := validator.SchemaValidator(incschema, inputBytes, inputFile, data)
	if err != nil { // return just the first error? TODO figure out error handling
		return cont, 0, errs[0]
	}

	// Take the input
	var inputFactory factory
	err = yaml.Unmarshal(inputBytes, &inputFactory)
	if err != nil {
		return cont, 0, fmt.Errorf("0002 %v when opening %s", err, inputFile)
	}
	if len(inputFactory.Create) == 0 {
		return cont, 0, fmt.Errorf("0003 No frames declared in %s", inputFile)
	}

	// set up the cont with frame useful contets
	frameCont := context.Background()
	credentials.InitDecoder(&frameCont, profile, httpKeys...)
	frameCont = gridgen.InitAliasBox(frameCont)

	holder := base{importedFactories: make(map[string]factory), importedWidgets: make(map[string]json.RawMessage),
		jsonFileLines: data, authBody: authDecoder, metadataParams: map[string][]string{},
		frameBase: frameCont,
	}

	baseDir := filepath.Dir(inputFile)
	err = holder.factoryInit(inputFactory, baseDir, "", []string{baseDir}, []int{})
	if err != nil {
		return cont, 0, err
	}

	// init a global alias map for use in grid gen
	aliasMap := SyncMap{make(map[string]string), &sync.Mutex{}}
	cont = context.WithValue(cont, aliasKey, aliasMap)
	// add the base factory and the imported schemas
	cont = context.WithValue(cont, updates, inputFactory)
	cont = context.WithValue(cont, frameHolders, holder)
	// add the working directory everything is relative to
	cont = context.WithValue(cont, factoryDir, filepath.Dir(inputFile))

	return cont, len(inputFactory.Create), nil
}

// Make a map of all the factories and their information
// so that are all open at once and there is no reopening of files and changing of data
func (b *base) factoryInit(jsonFactory factory, mainPath, parent string, factoryPaths []string, positions []int) error {

	if len(positions) > 30 {
		return fmt.Errorf("0004 recursive set initialisation file detected, the maximum dotpath depth of 30 has been reached")
	}

	for i, f := range jsonFactory.Include {

		// can we find the file
		fileBytes, path, err := FileSearch(b.authBody, f.URI, mainPath, factoryPaths)

		if err == nil {

			// check if the bytes have children by being a json factory
			var newF factory
			err := yaml.Unmarshal(fileBytes, &newF)
			if err != nil {
				return fmt.Errorf("0005 error parsing %s: %v", path, err)
			}

			if _, ok := b.importedWidgets[parent+f.Name]; ok {
				return fmt.Errorf("0006 the alias %s is repeated, every alias is required to be unique", parent+f.Name)
			} else if _, ok := b.importedFactories[parent+f.Name]; ok {
				return fmt.Errorf("0006 the alias %s is repeated, every alias is required to be unique", parent+f.Name)
			}

			// schema validation to sort between widgets and factories
			factLines := make(validator.JSONLines)
			err = validator.Liner(fileBytes, path, "factory", factLines) // treat it as a factory update
			if err != nil {
				return err
			}

			var validatorError error
			if err := validator.SchemaValidator(incschema, fileBytes, parent, factLines); err != nil {
				// @TODO include a better error handling method
				b.importedWidgets[parent+f.Name] = fileBytes
				validatorError = validator.Liner(fileBytes, path, "widget", b.jsonFileLines)
			} else {
				// schema check here?
				b.importedFactories[parent+f.Name] = newF
				validatorError = validator.Liner(fileBytes, path, "factory", b.jsonFileLines)
			}

			if validatorError != nil {

				return validatorError
			}

			// pass thorugh the factory as it won't run for length 0
			// append the path of where this was found
			parents := factoryPaths
			if !slices.Contains(parents, path) {
				// only append if its not an older path. As this will not effect the search order
				// search the depth of your tree. not the neighbours
				parents = append(parents, path)
			}

			err = b.factoryInit(newF, path, parent+f.Name+".", parents, append(positions, i))
			if err != nil { // return the error up the chain
				return err
			}
		} else {

			return err
			// fmt.Println(fmt.Errorf("Error opening %v:, %v\n", p, err))
		}

		b.metadataParams[parent+f.Name] = f.Args
	}

	return nil
}

// FileSearch searches for a factory file. Checking http sources, then http sources appended to the parent path.
// Then it searches local files, before appending the file path onto the parent paths.
func FileSearch(authBody credentials.Decoder, uri, mainPath string, parentPaths []string) (fileBytes []byte, folderFilePath string, fileErr error) {
	fileBytes, fileErr = authBody.Decode(uri)

	// generate the input path per run to stop overwriting errors

	if fileErr == nil {

		return fileBytes, uri, nil
	}

	inputPath, _ := url.JoinPath(mainPath, uri)
	// inputPath = filepath.Clean(filepath.Join(inputPath, f.URI))
	fileBytes, fileErr = authBody.Decode(inputPath)

	if fileErr == nil {
		return fileBytes, mainPath, nil
	}

	for _, path := range parentPaths {

		// Check relative to the mainjson
		inputPath, _ = filepath.Abs(filepath.Join(path, uri))
		// inputPath = filepath.Clean(filepath.Join(inputPath, f.URI))
		fileBytes, fileErr = os.ReadFile(inputPath)

		destFolder := filepath.Dir(inputPath)

		if fileErr == nil {
			return fileBytes, destFolder, nil
		}
	}

	// check relative to the location of the executable
	inputPath, _ = filepath.Abs(uri)
	// inputPath = filepath.Clean(filepath.Join(inputPath, f.URI))
	fileBytes, fileErr = os.ReadFile(inputPath)
	destFolder := filepath.Dir(inputPath)

	if fileErr == nil {
		return fileBytes, destFolder, nil
	}

	// finally check for OPENTSG_HOME
	TSGHome := os.Getenv("OPENTSG_HOME")
	if TSGHome != "" {
		// check for it
		inputPath, _ = filepath.Abs(filepath.Join(TSGHome, uri))
		// inputPath = filepath.Clean(filepath.Join(inputPath, f.URI))
		fileBytes, fileErr = os.ReadFile(inputPath)

		destFolder := filepath.Dir(inputPath)

		if fileErr == nil {
			return fileBytes, destFolder, nil
		}
	}

	// add searched locations
	return fileBytes, "", fileErr
}
