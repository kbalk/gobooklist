/*
Package booklist provides functions for searching a library's catalog website.

Issues the appropriate POST requests to search the catalog.  The requests
are in JSON format, as are the responses.  The request data contains
filters to narrow the search to an author, year and media type.  The
reponses are used to determine the number of publications matching those
filters and the list of publications.
*/
package booklist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/op/go-logging"
)

const (
	// Timeout in seconds for HTTP connect and read.
	timeout = 5

	// Maximum number of publications returned in a response.
	maxHitsPerPage = 30
)

var (
	// Used for 'cache busting'; in lieu of timezone information or
	// microsecond precision an increment is used.
	timestampIncrement int64 = 1

	// Current year as a string; used in filtering.
	yearFilter = time.Now().UTC().Format("2006")
)

// PublicationInfo provides the name and media type for a given publication.
type PublicationInfo struct {
	Media       string
	Publication string
}

// CatalogInfo provides the info needed to search for a given author and media.
type CatalogInfo struct {
	URL    string
	Author string
	Media  string
	Log    *logging.Logger
}

// facetFilter represents a map of filters used as POST JSON data.
type facetFilter map[string]string

// searchFilter is the collection of JSON data provided in a POST request.
type searchFilter struct {
	AddToHistory   bool          `json:"addToHistory"`
	DbCodes        []string      `json:"dbCodes"`
	HitsPerPage    int           `json:"hitsPerPage"`
	SortCriteria   string        `json:"sortCriteria"`
	StartIndex     int           `json:"startIndex"`
	TargetAudience string        `json:"targetAudience"`
	FacetFilters   []facetFilter `json:"facetFilters"`
	SearchTerm     string        `json:"searchTerm"`
}

// resourceInfo represents the map of resources returned in POST response.
type resourceInfo map[string]interface{}

// PublicationSearch coordinates the catalog search and parsing of the response.
//
// To perform a search on the library's catalog, two types of requests
// are needed:  one to retrieve the total number of publications
// available given a set of filters, the other to retreive publication
// information up to 'hitsPerPage' per request.
//
// These two requests are issued for the current year and again for
// publications with no known publication date.  In both cases, the search
// is also filtered for the given author with the given media type.
//
// Returns a list of tuples containing the media type and publication
// title for all publications in the current year or of an unknown year.
//
func (c CatalogInfo) PublicationSearch() ([]PublicationInfo, error) {
	if c.Author == "" || c.Media == "" {
		return nil, fmt.Errorf("arguments must be non-null:  "+
			"author=%s, media=%s'", c.Author, c.Media)
	}

	// Perform two sets of requests - one for publications within the
	// current year and one for publications of an unknown year.
	var filteredPubs []PublicationInfo
	var filters []facetFilter
	for _, year := range []string{"unknown", yearFilter} {
		filters = []facetFilter{
			facetFilter{
				"facetDisplay": year,
				"facetValue":   year,
				"facetName":    "Year",
			},
			facetFilter{
				"facetDisplay": c.Media,
				"facetValue":   c.Media,
				"facetName":    "Format",
			},
		}

		// Determine how many publications to expect so we know when
		// to stop issuing requests.
		totalCount, err := c.publicationsCount(filters)
		if err != nil {
			return nil, err
		}

		if totalCount == 0 {
			continue
		}

		// Loop issuing requests until all the publications have been
		// retrieved
		currentCount := 0
		for currentCount < totalCount {
			pubs, err := c.publications(filters)
			if err != nil {
				return nil, err
			}

			currentCount += len(pubs)
			c.Log.Debug("currentCount: %d", currentCount)

			// Apply additional filters that can't be handled in
			// POST request.
			c.applyLocalFilters(pubs, &filteredPubs)
		}

		// Retrieved more publications than expected?
		if currentCount > totalCount {
			return nil, fmt.Errorf("Received more publications "+
				"than expected; expected %d currently have %d",
				totalCount, currentCount)
		}
	}

	return filteredPubs, nil
}

// publicationsCount requests total number of publications for the given author.
func (c CatalogInfo) publicationsCount(filters []facetFilter) (int, error) {
	type hitResults struct {
		Success bool `json:"success"`
		Count   int  `json:"totalHits"`
	}
	results := new(hitResults)

	err := c.issueRequest("search/count", filters, &results)
	if err != nil {
		return 0, err
	}

	if !results.Success {
		return 0, fmt.Errorf("failed to retrieve total number " +
			"of matches on author, media and year")
	}

	c.Log.Debugf("Expected number of matches:  %d", results.Count)
	return results.Count, nil
}

// publications requests a page of publications for the given author.
func (c CatalogInfo) publications(filters []facetFilter) ([]resourceInfo, error) {
	type searchResults struct {
		totalHits    int
		facetFilters []facetFilter
		Resources    []resourceInfo `json:"resources"`
	}
	results := new(searchResults)

	err := c.issueRequest("search", filters, &results)
	if err != nil {
		return nil, err
	}
	c.Log.Debugf("Number of resources found: %d", len(results.Resources))
	return results.Resources, nil
}

// applyLocalFilters applies additional localized filters on publications
//
// Filter more precisely on the author name as the search can sometimes
// retrieve other publications that are not from the author.  Also,
// as the author could be one of several authors for the publication,
// an exact match shouldn't be performed on the name.
//
// Additionally, check for missing dictionary values for title and
// media type and use 'Unknown' as a replacement.
//
func (c CatalogInfo) applyLocalFilters(pubs []resourceInfo, filteredResults *[]PublicationInfo) {
	for _, publication := range pubs {
		// Some books don't have authors - don't know why,
		// but 'The Mystery Writers of America cookbook' is one
		// of them; it shows up in a search for Sue Grafton.
		var author string
		var ok bool
		if author, ok = publication["shortAuthor"].(string); !ok {
			continue
		}

		var format string
		if format, ok = publication["format"].(string); !ok {
			format = "Unknown"
		}

		var title string
		if title, ok = publication["shortTitle"].(string); !ok {
			title = "Unknown"
		}

		if c.Author == author {
			c.Log.Debugf("media:  %s, title:  %s", format, title)
			*filteredResults = append(*filteredResults, PublicationInfo{
				Media:       format,
				Publication: title,
			})
		}
	}
}

// issueRequest issues a post request and checks for an error in the response.
func (c CatalogInfo) issueRequest(endpt string, filters []facetFilter, target interface{}) error {

	// Create the url that includes the given endpoint and add the
	// 'cache buster' timestamp parameter.
	u, err := url.Parse(c.URL + endpt)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("_", makeTimestamp())
	u.RawQuery = params.Encode()

	// Create the POST's json data containing the filters, sort and other
	// info.
	search := searchFilter{
		AddToHistory: true,
		HitsPerPage:  maxHitsPerPage,
		SortCriteria: "NewlyAdded",
		FacetFilters: filters,
		SearchTerm:   c.Author,
	}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(search)
	if err != nil {
		return fmt.Errorf("unable to decode filters: %#v, error: %s",
			search, err)
	}

	// Formulate the POST request with specific header values and a timeout
	// value.  The POST request will contain the search filter in json
	// format.
	req, err := http.NewRequest("POST", u.String(), b)

	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Ls2pac-config-type", "pac")
	req.Header.Set("Ls2pac-config-name", "default - Go Live load")
	req.Header.Set("Referer", c.URL)

	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		if err != nil {
			return fmt.Errorf("POST request '%s' failed; %s", u, err)
		} else {
			return fmt.Errorf("POST request '%s' failed; "+
				"HTTP error: %s",
				u, http.StatusText(resp.StatusCode))
		}
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		return fmt.Errorf("unable to decode response to '%s': "+
			"response %#v, error: %s", c.URL, resp, err)
	}
	return nil
}

// Return a 13-digit timestamp; used as a 'cache buster' in requests.
//
// With the CARL.X system, the parameter '_' in a request appears to
// contain a value used as a 'cache buster'.  A cache buster value
// is checked to see if it's different from a prior request's value
// and if so, then new data is retrieved rather than using cached data.
//
// As the CARL.X system uses a 13-digit timestamp, so will we.  To get
// 13-digits, we multiply a timestamp by 1000.  That yields zeros at the
// end of the number, so we add an increment to the end to keep successive
// requests unique.
func makeTimestamp() string {
	timestampIncrement++
	utcTime := time.Now().UTC().UnixNano()
	timestamp := utcTime / (int64(time.Millisecond) / int64(time.Nanosecond))
	return fmt.Sprintf("%d", timestamp+timestampIncrement)
}
