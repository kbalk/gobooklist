/*
Contains Configurator and ConfigError classes

The Configurator class reads and validates the configuration file.  The
configuration file contains:

    - URL for the library's catalog website,
    - optional default for desired media type ('Book' is assumed otherwise),
    - list of authors; each entry in the list must contain the author's first
      and last name, and optionally the desired media type for that author.

The ConfigError class is used to report errors found in the configuration
file.

This particular implementation expects a config file in YAML format.  The
tags are as follows:

    catalog-url:
	Required.  Must be a valid URL for a website using the CARL.X
	Integrated Library System.
    media-type:
	Optional.  The default media type is 'book'; allowed types are:
	    book
	    electronic resource
	    ebook
	    eaudiobook
	    book on cd
	    large print
	    music cd
	    dvd
	    blu-ray
	These types can be in upper, lower or mixed case as the case will
	be ignored.  Types that are more than one word can be enclosed in
	quotes or not; it won't matter.
	Note that some media types are supersets, i.e., a type of 'book'
	includes 'large print' books.  A type of 'electronic resource'
	includes 'ebook'.
    authors:
	Required.  List of authors specified by first and last name and
	optionally by media-type.
    authors sub-tags:
	firstname:
	    Required.  First name of author.
	lastname:
	    Required.  Last name of author.
	media-type:
	    Optional.  See media-type above for the allowed values.

Example YAML config file

    catalog-url: https://catalog.library.loudoun.gov/
    media-type: Book
    authors:
        - firstname: James
          lastname: Patterson
          media-type: book on cd
        - firstname: Alexander
          lastname: McCall Smith
*/
package booklist

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

const (
	// DefaultMediaType is the default media type.
	DefaultMediaType = "Book"
)

// Config is the high level structure for the YAML config file.
type Config struct {
	URL     string       `yaml:"catalog-url"`
	Media   string       `yaml:"media-type,omitempty"`
	Authors []AuthorInfo `yaml:"authors,flow"`
}

// AuthorInfo provides the sub fields for the Authors field.
type AuthorInfo struct {
	Firstname string
	Lastname  string
	Media     string `yaml:"media-type,omitempty"`
}

// schema is the schema for the YAML configuration file.
var schema = `
{
        "$schema": "http://json-schema.org/draft-04/schema#",
        "type": "object",
        "required": ["URL", "Authors"],
        "properties": {
            "URL": {"type": "string", "format": "uri"},
            "Media": {"type": "string", "format": "media"},
            "Authors": {
                "type": "array",
                "items": {
                    "type": "object",
                    "required": ["Firstname", "Lastname"],
                    "properties": {
                        "Firstname": {"type": "string", "minLength": 1},
                        "Lastname": {"type": "string", "minLength": 1},
                        "Media": {"type": "string", "format": "media"}
                    }
                }
            }
        },
        "additionalProperties": false
}`

// MediaTypes - array of supported media types.
//
// The following list contains most of the supported media types
// allowed by the CARL-X ILS.  This is a map with a media type config name
// as a key and the equivalent name for use in the URL query string as the
// value.
//
// Note:  when validating the media type name found in the config file,
// the name will first be converted to lower case before comparing it
// against this list.
var MediaTypes = map[string]string{
	"book":                "Book",
	"electronic resource": "Electronic Resource",
	"ebook":               "eBook",
	"eaudiobook":          "eAudioBook",
	"book on cd":          "Book on CD",
	"large print":         "Large Print",
	"music cd":            "Music CD",
	"dvd":                 "DVD",
	"blu-ray":             "Blu-Ray",
}

// mediaFormatChecker specifies a custom format type, 'media' to gojsonschema.
type mediaFormatChecker struct{}

// IsFormat provides the logic to validate the custom format type of 'media'.
func (f mediaFormatChecker) IsFormat(input string) bool {
	if input == "" {
		return true
	}
	_, ok := MediaTypes[strings.ToLower(input)]
	return ok
}

// convertMediaType converts media type fields to values needed by URL request.
//
// Note:  this assumes the config file has already been validated.
func convertMediaType(config *Config) {
	if config.Media != "" {
		config.Media = MediaTypes[strings.ToLower(config.Media)]
	}
	for i := range config.Authors {
		lcMediaType := strings.ToLower(config.Authors[i].Media)
		if lcMediaType != "" {
			config.Authors[i].Media = MediaTypes[lcMediaType]
		}
	}
}

// ReadConfig return contents of file into a byte slice if readable file exists.
func ReadConfig(configFileName string) ([]byte, error) {
	path, err := filepath.Abs(configFileName)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is a directory; must be a file",
			configFileName)
	}

	return ioutil.ReadFile(path)
}

// ValidateConfig validates the YAML file contents against a schema.
func ValidateConfig(in []byte) (Config, error) {
	var config Config

	if len(in) == 0 {
		return config, fmt.Errorf("configuration content is empty")
	}

	// Marshal the contents of the YAML into the Go structure, 'config'.
	err := yaml.Unmarshal(in, &config)
	if err != nil {
		return config,
			fmt.Errorf("unable to parse YAML config file:  %s", err)
	}
	//TBD
	fmt.Println(config)

	// To prepare for validation, load the config structure, add the
	// media format checker to the schema, then load the schema.
	structLoader := gojsonschema.NewGoLoader(config)

	gojsonschema.FormatCheckers.Add("media", mediaFormatChecker{})
	schemaLoader := gojsonschema.NewStringLoader(schema)

	// Validate the config structure against the schema.
	result, err := gojsonschema.Validate(schemaLoader, structLoader)

	// Any problems with the schema itself?
	if err != nil {
		return config, fmt.Errorf("invalid schema: %s", err)
	}

	// Any validation issues?  If so, create an array of the validation
	// errors.  Unfortunately, this implementation of schema validation
	// doesn't provide line numbers where errors are found.
	if !result.Valid() {
		var errmsg []string
		for _, err := range result.Errors() {
			errmsg = append(errmsg, fmt.Sprintf("- %s\n", err))
		}
		return config, fmt.Errorf("YAML failed schema validation: %s",
			strings.Join(errmsg[:], "\n"))
	}

	// Transform the media types to the values needed for the URL request.
	convertMediaType(&config)

	return config, err
}

// Stringer function for Config struct.
func (config Config) String() string {
	var authorInfo []string
	var line string

	for _, info := range config.Authors {
		if info.Media != "" {
			line = fmt.Sprintf("   %v %v; %s",
				info.Firstname, info.Lastname, info.Media)
		} else {
			line = fmt.Sprintf("   %v %v", info.Firstname,
				info.Lastname)
		}
		authorInfo = append(authorInfo, line)
	}

	return fmt.Sprintf("%v\n%v\n%s\n", config.URL, config.Media,
		strings.Join(authorInfo, "\n"))
}
