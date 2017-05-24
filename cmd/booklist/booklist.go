/* Main package for booklist; lists author publications in current year.

Automates the search of a public library website to retrieve publications
for a configured list of authors.

Takes a YAML file with the library's catalog URL and list of authors as
input.  Issues the appropriate requests to that library's website for the
latest publications for those authors.  Only works for libraries using
the CARL.X Integrated Library System (ILS).

The CARL.X ILS is based on Web 2.0 technologies so its possible to issue
the POST requests needed to obtain the number of expected publications
and the list of those publications.  It has an "open" API, but that API
appears to be open only to paying customers, i.e., the library staff.

A request URL to search for a given book consists of an array of
facetFilters.  You can filter on format (media type), publication year,
new titles, etc.  The 'New Titles' filter permits finer granularity in
time, e.g., weeks or months.  However, if you go to a library website
using CARL.X and manually select a format filter, the 'New Titles' filter
is no longer selectable.  It's not clear why that is so, but a 'Publication
Year' filter is still permitted.

This tool defaults to a search within a publication year and that year is
the current one.  Media with an unknown publication time period will also
be returned from a search as they are future releases that might be
available in the current year.

Usage: booklist [-h] [-d] config_file
    Search a public library's catalog website for this year's publications
    from authors listed in the given config file.

    positional arguments:
      config_file  Config file containing catalog url and list of authors
    optional arguments:
      -h, --help   show this help message and exit
      -d, --debug  Print debug information to stderr
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kbalk/gobooklist/booklist"
	"github.com/op/go-logging"
)

// initLogging initializes the format and debug level for the stderr logging.
func initLogging(log *logging.Logger, debug bool) {
	stderrLog := logging.NewLogBackend(os.Stderr, "", 0)

	format := logging.MustStringFormatter(
		`%{time:15:04:05} %{level:.5s} %{message}`)
	logFormatter := logging.NewBackendFormatter(stderrLog, format)

	logLevel := logging.AddModuleLevel(logFormatter)
	level := logging.ERROR
	if debug {
		level = logging.DEBUG
	}
	logLevel.SetLevel(level, "")

	log.SetBackend(logLevel)
}

// Retrieve and print the author publications for current year.
func printSearchResults(config booklist.Config, log *logging.Logger) error {
	// The default type is the value specified in the config file or
	// if not found, the standard default type.
	defaultMedia := booklist.DefaultMediaType
	if config.Media == "" {
		defaultMedia = config.Media
	}

	for _, authorInfo := range config.Authors {
		authorName := fmt.Sprintf("%s, %s",
			authorInfo.Lastname, authorInfo.Firstname)

		media := defaultMedia
		if authorInfo.Media != "" {
			media = authorInfo.Media
		}

		fmt.Printf("%s -- %ss:\n", authorName, media)
		c := booklist.CatalogInfo{
			URL:    config.URL,
			Author: authorName,
			Media:  media,
			Log:    log,
		}
		results, err := c.PublicationSearch()
		if err != nil {
			return err
		}
		if results == nil {
			continue
		}

		// Print the search results; each entry in the results list
		// is a tuple containing the media type and publication name
		// (e.g., book title).  Since some media types are supersets
		// of other media types, it seemed useful to provide that
		// extra information.
		maxWidth := 0
		for _, info := range results {
			l := len(info.Media)
			if l > maxWidth {
				maxWidth = l
			}
		}
		for _, pubInfo := range results {
			fmt.Printf("  [%-*s]  %s\n",
				maxWidth, pubInfo.Media, pubInfo.Publication)
		}
	}
	return nil
}

// main processes command line args then retrieve search results from library.
func main() {
	flag.Usage = func() {
		usageText := `Usage: go_booklist: [-h] [-d] config_file

  Search a public library's catalog website for this year's publications
  from authors listed in the given config file.

  config_file
        Config file with library's catalog url and list of authors`
		fmt.Fprintln(os.Stderr, usageText)
		flag.PrintDefaults()
	}
	var debugFlag = flag.Bool("debug", false,
		"Print debug information to stderr")
	flag.Parse()

	// Verify that only one argument is supplied, that argument being
	// the configuration file.
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr,
			"ERROR:  config filename is required argument.\n\n")
		flag.Usage()
		os.Exit(1)
	}
	configFileName := flag.Arg(0)

	// Initialize logging for error and/or debug messages to stderr.
	var log = logging.MustGetLogger("booklist")
	initLogging(log, *debugFlag)

	// Verify the config exists and is readable, then read the contents.
	configBytes, ok := booklist.ReadConfig(configFileName)
	if ok != nil {
		log.Error(ok)
		os.Exit(1)
	}

	// Validate the config file contents and retrieve the parsed results.
	config, ok := booklist.ValidateConfig(configBytes)
	if ok != nil {
		log.Error(ok)
		os.Exit(1)
	}
	log.Debug(config)

	// Retrieve the publications for the authors in the configuration file
	// and print the results.
	if err := printSearchResults(config, log); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	os.Exit(0)
}
