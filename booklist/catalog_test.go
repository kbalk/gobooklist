// Unit tests related to catalog search. //
package booklist

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/op/go-logging"
)

var testLog *logging.Logger

func init() {
	testLog = logging.MustGetLogger("catalog_test")

	testBackendLog := logging.NewLogBackend(ioutil.Discard, "", 0)
	logLevel := logging.AddModuleLevel(testBackendLog)
	logLevel.SetLevel(0, "")

	testLog.SetBackend(logLevel)
}

func TestLiveGoodSearch(t *testing.T) {
	t.Log("test search using good configuration file and real connection.")
	// Note:  Because this is a live search, it could fail if sometime
	// in the future the library removes the expected books from their
	// inventory.  The probability of that happening is reduced by
	// specifying a year that's not too far in the past and using a
	// popular author.
	expected := []PublicationInfo{
		{"Large Print", "J is for judgment"},
		{"Large Print", "K is for killer : a Kinsey Millhone mystery"},
		{"Large Print", "L is for lawless"},
		{"Large Print", "M is for malice : a Kinsey Millhone mystery"},
		{"Large Print", "N is for noose a Kinsey Millhone mystery"},
		{"Large Print", "O is for outlaw"},
		{"Book", "X"},
		{"Large Print", "X"},
	}

	liveURL := "https://catalog.library.loudoun.gov/"
	c := CatalogInfo{
		URL:    liveURL,
		Author: "Grafton, Sue",
		Media:  "Book",
		Year:   "2015",
		Log:    testLog,
	}
	pubInfo, err := c.PublicationSearch()
	if err != nil {
		t.Errorf("Test of live search at '%s' failed: %s.", liveURL, err)
	}

	for i, info := range pubInfo {
		if info.Media != expected[i].Media {
			t.Errorf("Expected media of %s, got %s.", info.Media,
				expected[i].Media)
		}
		if info.Publication != expected[i].Publication {
			t.Errorf("Expected publication of %s, got %s.",
				info.Publication, expected[i].Publication)
		}
	}
}

func TestBadURL(t *testing.T) {
	t.Log("Bad URL for libary website.")
	c := CatalogInfo{
		URL:    "http:/nosuchurl.com",
		Author: "Grafton, Sue",
		Media:  "Book",
		Year:   CurrentYear,
		Log:    testLog,
	}
	_, err := c.PublicationSearch()

	expectedErrMsg := ""
	if err == nil {
		t.Errorf("Failed to reject bad URL.")
	} else if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s'; got %s.",
			expectedErrMsg, err)
	}
}

func TestMissingInfo(t *testing.T) {
	t.Log("Missing media in catalog information")
	goodURL := "https://catalog.library.loudoun.gov/"
	author := "Grafton, Sue"
	testCases := []struct {
		URL    string
		media  string
		author string
		year   string
		msg    string
	}{
		{"", "Book", author, CurrentYear, "url"},
		{goodURL, "", author, CurrentYear, "media"},
		{goodURL, "Book", "", CurrentYear, "author"},
		{goodURL, "Book", author, "", "year"},
	}
	for _, tc := range testCases {
		c := CatalogInfo{
			URL:    tc.URL,
			Author: tc.author,
			Media:  tc.media,
			Year:   tc.year,
			Log:    testLog,
		}
		_, err := c.PublicationSearch()

		expectedErrMsg := "catalog information must be non-null"
		if err == nil {
			t.Errorf("Missing %s field not detected", tc.msg)
		} else if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("Expected error message for missing %s to "+
				"contain '%s'; got %s",
				tc.msg, expectedErrMsg, err)
		}
	}
}
