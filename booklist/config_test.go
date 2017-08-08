// Unit tests related to configuration validation. //
package booklist

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestGoodConfig(t *testing.T) {
	t.Log("test a good configuration file.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: Book
        authors:
            - firstname: Sue
              lastname:  Grafton
              media-type: eBook

            - firstname: Stephan
              lastname:  King
        `
	config, ok := ValidateConfig([]byte(configString))
	if ok != nil {
		t.Errorf("Schema should be valid; instead got error: %s.", ok)
	}

	if config.URL != "https://catalog.library.loudoun.gov/" {
		t.Errorf("Expected URL to be "+
			"'https://catalog.library.loudoun.gov/', got %s.",
			config.URL)
	}
	if config.Media != "Book" {
		t.Errorf("Expected default media type to be 'Book'; got %s.",
			config.Media)
	}
	if config.Authors[0].Firstname != "Sue" &&
		config.Authors[0].Lastname != "Grafton" {
		t.Errorf("Expected first author to be 'Sue Grafton' ; got: '%s %s'.",
			config.Authors[0].Firstname, config.Authors[0].Lastname)
	}
	if config.Authors[0].Media != "eBook" {
		t.Errorf("Expected first author's media type to be "+
			"'eBook'; got '%s'.", config.Authors[0].Media)
	}
	if config.Authors[1].Firstname != "Stephan" &&
		config.Authors[1].Lastname != "King" {
		t.Errorf("Expected second author to be 'Stephan King' ; got: '%s %s'.",
			config.Authors[1].Firstname, config.Authors[1].Lastname)
	}
	if config.Authors[1].Media != "" {
		t.Errorf("Expected second author's media type to be "+
			"unspecified; got '%s'.", config.Authors[1].Media)
	}
}

func TestNonexistentConfig(t *testing.T) {
	t.Log("attempt to read a non-existent config file.")
	var ok error

	_, ok = ReadConfig("file_does_not_exist")
	if ok == nil {
		t.Error("Expected error due to non-existent config file.")
	}
	if !strings.Contains(ok.Error(), "cannot find the file") {
		t.Errorf("Expected error message to contain "+
			"'cannot find the file'; got: %s.", ok)
	}
}

func TestEmptyConfig(t *testing.T) {
	t.Log("attempt to read an empty config file.")
	tmpfile, err := ioutil.TempFile("", "tmpfile_empty_config")
	tmpfileName := tmpfile.Name()

	if err != nil {
		t.Errorf("Unable to create temp file '%s' for unit test.",
			tmpfileName)
	}
	defer os.Remove(tmpfileName)

	if err := tmpfile.Close(); err != nil {
		t.Errorf("Error closing temp file '%s' for unit test.",
			tmpfileName)
	}

	var ok error
	var content []byte

	content, ok = ReadConfig(tmpfileName)
	if ok != nil {
		t.Error("Empty config file should not be an error until " +
			"contents are processed.")
	}

	_, ok = ValidateConfig(content)
	if ok == nil {
		t.Error("Expected error due to empty config file as input.")
	}
	if !strings.Contains(ok.Error(), "content is empty") {
		t.Errorf("Expected error message to contain "+
			"'content is empty'; got: %s.", ok)
	}
}

func TestFileAsDir(t *testing.T) {
	t.Log("attempt to read a directory instead of a config file")

	_, ok := ReadConfig(".")
	if ok == nil {
		t.Error("Config file cannot be a directory; must be a file.")
	}
	if !strings.Contains(ok.Error(), "must be a file") {
		t.Errorf("Expected error message to contain must be a file'; "+
			"got: %s.", ok)
	}
}

func TestBadYamlFormat(t *testing.T) {
	t.Log("non-YAML file as input")
	var ok error
	var content []byte

	content, ok = ReadConfig("config.go")
	_, ok = ValidateConfig(content)
	if ok == nil {
		t.Errorf("Expected error with non-YAML file (config.go) as input.")
	}
	if !strings.Contains(ok.Error(), "unable to parse YAML config file") {
		t.Errorf("Expected error message to contain "+
			"'unable to parse YAML config file'; got: %s.", ok)
	}
}

func TestExtraneousSpaces(t *testing.T) {
	t.Log("good YAML content with spaces before and after values.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: Book
        authors:
            - firstname:          Sue
              lastname:   Grafton
              media-type:      eBook

            - firstname:                  Stephan
              lastname:      King
        `
	config, ok := ValidateConfig([]byte(configString))
	if ok != nil {
		t.Errorf("Schema should be valid; instead got error: %s.", ok)
	}
	if config.Authors[0].Firstname != "Sue" {
		t.Errorf("Expected first name to contain no spaces; got: '%s'.",
			config.Authors[0].Firstname)
	}
}

func TestQuotedText(t *testing.T) {
	t.Log("all spaces in firstname field.")
	// Put the full name into the 'lastname' field and use a space for
	// the 'firstname' field.  The space in the firstname is actually kept
	// and not stripped; if a space is not used the schema will complain
	// as the minimum length for the field is 1.  I'm not sure why the space
	// isn't stripped.  The quoted lastname is fine.
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: Book
        authors:
            - firstname: ' '
              lastname: 'Sue Grafton'

            - firstname: M.C.
              lastname: Quotes missing
        `
	config, ok := ValidateConfig([]byte(configString))
	if ok != nil {
		t.Errorf("Schema should be valid; instead got error: %s.", ok)
	}
	if config.Authors[0].Firstname != " " {
		t.Errorf("Expected first name to contain a space; got: '%s'.",
			config.Authors[0].Firstname)
	}
}

func TestInvalidURLs(t *testing.T) {
	t.Log("Tests of various invalid or missing URLs.")
	testCases := []struct {
		description string
		url         string
	}{
		{"null url", ""},
		{"single word url", "        catalog-url: badurl"},
		{"no domain in url", "        catalog-url: catalog.library.loudoun.gov"},
	}

	configString := `
        media-type: Book
        authors:
           - firstname: Sue
             lastname: Grafton
        `

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, ok := ValidateConfig([]byte(tc.url + configString))
			if ok == nil {
				t.Errorf("Schema validation of config file should " +
					"fail due to bad URL.")
			}
			if !strings.Contains(ok.Error(), "uri") {
				t.Errorf("Expected error message to contain 'uri' "+
					"for url of %s; got: %s.", tc.url, ok)
			}
		})
	}
}

func TestURLWithoutTrailingSlash(t *testing.T) {
	t.Log("Test URL with missing trailing slash")
	const configString = `
        catalog-url: https://junk.com
        media-type: book
        authors:
            - firstname: Stephan
              lastname:  King
        `
	config, ok := ValidateConfig([]byte(configString))
	if ok != nil {
		t.Errorf("Schema validation of config file should " +
			"not fail due to missing trailing slash is URL.")
	}
	if !strings.HasSuffix(config.URL, "/") {
		t.Errorf("Trailing slash not ended to end of URL; expected "+
			"%s, got %s", config.URL, config.URL+"/")
	}
}

func TestBadMediaType(t *testing.T) {
	t.Log("Bad default media type.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: nonsense
        authors:
            - firstname: Sue
              lastname:  Grafton
              media-type: eBook

            - firstname: Stephan
              lastname:  King
        `
	_, ok := ValidateConfig([]byte(configString))
	if ok == nil {
		t.Errorf("Schema validation of config file should " +
			"fail due to bad default media type.")
	}
	if !strings.Contains(ok.Error(), "Media: Does not match") {
		t.Errorf("Expected error message to contain "+
			"'Media: Does not match'; got: %s.", ok)
	}
}

func TestBadAuthorMediaType(t *testing.T) {
	t.Log("Bad default media type.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: Book
        authors:
            - firstname: Sue
              lastname:  Grafton
              media-type: nonsense

            - firstname: Stephan
              lastname:  King
        `
	_, ok := ValidateConfig([]byte(configString))
	if ok == nil {
		t.Errorf("Schema validation of config file should " +
			"fail due to bad default media type.")
	}
	if !strings.Contains(ok.Error(), "Media: Does not match") {
		t.Errorf("Expected error message to contain "+
			"'Media: Does not match'; got: %s.", ok)
	}
}

func TestOptionalMediaType(t *testing.T) {
	t.Log("No media type specified at all; this is allowed.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        authors:
            - firstname: Sue
              lastname:  Grafton

            - firstname: Stephan
              lastname:  King
        `
	_, ok := ValidateConfig([]byte(configString))
	if ok != nil {
		t.Errorf("Schema validation of config file should " +
			"not fail due to missing media types.")
	}
}

func TestMediaTypeTransformation(t *testing.T) {
	t.Log("Verify all the media types are correctly converted.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: XXX
        authors:
            - firstname: Sue
              lastname:  Grafton
              media-type: XXX
        `

	for mediaType, filterType := range MediaTypes {
		newString := strings.Replace(configString, "XXX", mediaType, -1)
		config, ok := ValidateConfig([]byte(newString))
		if ok != nil {
			t.Error("Schema validation of config file should " +
				"not fail due when using valid media types.")
		}
		if config.Media != filterType {
			t.Errorf("Expected conversion of media type '%s' to "+
				"yield '%s'", mediaType, filterType)
		}
		if config.Authors[0].Media != filterType {
			t.Errorf("Expected conversion of author's media type "+
				"'%s' to yield '%s'", mediaType, filterType)
		}
	}
}

func TestNoAuthors(t *testing.T) {
	t.Log("No authors.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: Book
        authors:
        `
	_, ok := ValidateConfig([]byte(configString))
	if ok == nil {
		t.Errorf("Schema validation of config file should " +
			"fail due to missing authors list.")
	}
	if !strings.Contains(ok.Error(), "Authors: Invalid type") {
		t.Errorf("Expected error message to contain "+
			"'Authors: Invalid type'; got: %s.", ok)
	}
}

func TestInvalidAuthorNames(t *testing.T) {
	t.Log("Tests of various invalid author names.")
	testCases := []struct {
		description string
		name        string
	}{
		{"no first or last name", ""},
		{"no last name", "lastname: Grafton"},
		{"no first name", "firstname: Sue"},
	}

	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: Book
        authors:
          - %s

          - firstname: Stephen
            lastname: King
        `

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, ok := ValidateConfig([]byte(fmt.Sprintf(configString, tc.name)))
			if ok == nil {
				t.Errorf("Schema validation of config file should " +
					"fail due to author list.")
			}
			if !strings.Contains(ok.Error(), "Author") {
				t.Errorf("Expected error message to contain 'Author; "+
					"for bad name of '%s'; got: %s.", tc.name, ok)
			}
		})
	}
}

func TestConfigStringer(t *testing.T) {
	t.Log("test config stringer function.")
	const configString = `
        catalog-url: https://catalog.library.loudoun.gov/
        media-type: Book
        authors:
            - firstname: Sue
              lastname:  Grafton
              media-type: eBook

            - firstname: Stephan
              lastname:  King
        `
	config, _ := ValidateConfig([]byte(configString))
	configStr := fmt.Sprintf("%s", config)

	const expectedStr = `https://catalog.library.loudoun.gov/
Book
   Sue Grafton; eBook
   Stephan King
`
	if configStr != expectedStr {
		t.Errorf("Expected config stringer to produce:%s\ngot\n%s",
			expectedStr, configStr)
	}
}
