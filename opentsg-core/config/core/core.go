// Package Core is used for handling factory objects for imports and frame generation
package core

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/validator"
	"github.com/mrmxf/opentsg-modules/opentsg-core/credentials"
)

////////////////
// This is a base of useful things that are used across the config files. Will be designed to be something used by each one

type testKey string

const (
	updates testKey = "update key for the array of objects"
	baseKey testKey = "base key for widgets"
	//	widgetbases  = "widget bases", 2}
	frameHolders    testKey = "The key for holding all the generated json"
	aliasKey        testKey = "base for aliases to run through out the program"
	aliasKeyBox     testKey = "base for aliases to run through out the program, but with the box method"
	lines           testKey = "the holder of the hashes of the name+content for line numbers and files"
	addedWidgets    testKey = "the key to access the list of added widgets to find missed aliases"
	factoryDir      testKey = "the directory of the main widget factory and everything is relative to"
	credentialsAuth testKey = "the holder of all the auth information provided by the user for accessing http sources"
	poskey          testKey = "the key that holds the frame position"
)

// then the rest is additions to the alias

type factory struct {
	Include  []factoryarr                `json:"include,omitempty" yaml:"include,omitempty"`
	Args     []arguments                 `json:"args" yaml:"args"`
	Create   []map[string]map[string]any `json:"create" yaml:"create"`
	Generate []generate                  `json:"generate" yaml:"generate"`
	// ADD a middleawre section here, may be difficult to keep tabs on
}

type arguments struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

type factoryarr struct {
	URI  string `json:"uri,omitempty" yaml:"uri,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Args []string `json:"args" yaml:"args"`
}

type generate struct {
	Name   []map[string]string            `json:"name" yaml:"name"` // map[string]any add the error handling later
	Range  []string                       `json:"range" yaml:"range"`
	Action map[string]map[string][]string `json:"action" yaml:"action"` // target(s)  data  updates
}

//go:embed jsonschema/includeschema.json
var incschema []byte

//go:embed jsonschema/widgetEssentials.json
var propsSchema []byte

// Processing structs
type base struct {
	authBody              credentials.Decoder
	frameBase             context.Context
	jsonFileLines         validator.JSONLines
	importedFactories     map[string]factory
	importedWidgets       map[string]json.RawMessage
	generatedFrameWidgets map[string]widgetContents
	metadataParams        map[string][]string
	metadataBucket        map[string]map[string]any
}

type widgetContents struct {
	Data        json.RawMessage
	Pos         int
	arrayPos    []int
	Widget      bool
	ColourSpace colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
}

// WidgetContents contains the raw widget properties
type WidgetContents struct {
	Data     json.RawMessage
	Pos      int
	arrayPos []int
	WidgetEssentials
}

// AliasIdentity is the name and zposition of a widget. Where zposition is the widgets position in the global array of widgets.
// As well as any other properties associated with the widget
type AliasIdentityHandle struct {
	ZPos     int
	FullName string
	Contents json.RawMessage
	WidgetEssentials
}

// AliasIdentity is the name and zposition of a widget. Where zposition is the widgets position in the global array of widgets.
// As well as any other properties associated with the widget
// This is the legacy version
type AliasIdentity struct {
	FullName    string
	ZPos        int
	WType       string
	Location    string
	GridAlias   string
	ColourSpace colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
}

// GetFrameWidgets returns a map of all the widgets and their properties
// This is the legacy version
func GetFrameWidgets(c context.Context) map[string]widgetContents {

	return c.Value(baseKey).(map[string]widgetContents)
}

// GetFrameWidgets returns  a map of all the widgets and their properties
// this contains the widgets "props" properties.
func GetFrameWidgetsHandle(c context.Context) map[string]WidgetContents {

	return c.Value(baseKey).(map[string]WidgetContents)
}

// SyncMap  is a map with a sync.Mutex to prevent concurrent writes.
type SyncMap struct {
	Data map[string]string
	Mu   *sync.Mutex
}

// GetAliasMap returns a syncMap that contains all the widget names that have been assigned an alias
func GetAliasMap(c context.Context) SyncMap {

	return c.Value(addedWidgets).(SyncMap)
}

// GetJSONLines returns the hash map of all the imported files and their lines.
// This is for use in conjunction with the validator package
func GetJSONLines(c context.Context) validator.JSONLines {

	return c.Value(lines).(validator.JSONLines)
}

// GetDir returns the directory that the base factory resides in.
func GetDir(c context.Context) string {
	s, ok := c.Value(factoryDir).(string)
	if !ok {
		s, _ := os.Getwd()

		return s
	}

	return s
}

// GetFramePosition returns the frame number of openTSG.
func GetFramePosition(c context.Context) int {
	pos := c.Value(poskey)
	if pos != nil {
		p, ok := pos.(int)
		if ok {
			return p
		}
	}

	return 0
}

// GetWebBytes is a wrapper of `credentials` where the configuration body is stored in config.
// This is to prevent several intialisations of the authbody or the data being passed around.
func GetWebBytes(c *context.Context, uri string) ([]byte, error) {

	var ok = false
	var auth credentials.Decoder

	if c != nil {
		auth, ok = (*c).Value(credentialsAuth).(credentials.Decoder)
	}
	// if there is not an authorisation body make a new one with no credentials
	if !ok {
		var err error
		auth, err = credentials.AuthInit("")
		if err != nil {
			return nil, err
		}
	}

	return auth.Decode(uri)
}
