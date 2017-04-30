// Main package for booklist; lists author publications in current year.
//
// Automates the search of a public library website to retrieve publications
// for a configured list of authors.
//
// Takes a YAML file with the library's catalog URL and list of authors as
// input.  Issues the appropriate requests to that library's website for the
// latest publications for those authors.  Only works for libraries using
// the CARL.X Integrated Library System (ILS).
//
// The CARL.X ILS is based on Web 2.0 technologies so its possible to issue
// the POST requests needed to obtain the number of expected publications
// and the list of those publications.  It has an "open" API, but that API
// appears to be open only to paying customers, i.e., the library staff.
//
// A request URL to search for a given book consists of an array of
// facetFilters.  You can filter on format (media type), publication year,
// new titles, etc.  The 'New Titles' filter permits final granularity in
// time, e.g., weeks or months.  However, if you go to a library website
// using CARL.X and manually select a format filter, the 'New Titles' filter
// is no longer selectable.  It's not clear why that is so, but a 'Publication
// Year' filter is still permitted.
//
// This tool defaults to a search within a publication year and that year is
// the current one.  Media with an unknown publication time period will also
// be returned from a search as they are future releases that might be
// available in the current year.
//
// Usage: go_booklist [-h] [-d] config_file
//     Search a public library's catalog website for this year's publications
//     from authors listed in the given config file.
//
//     positional arguments:
//       config_file  Config file containing catalog url and list of authors
//     optional arguments:
//       -h, --help   show this help message and exit
//       -d, --debug  Print debug information to stderr
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kbalk/gobooklist/booklist"
	"github.com/op/go-logging"
)

// initLogging initializes the format and debug level for the stderr logging.
func initLogging(debug bool) {
	stderrLog := logging.NewLogBackend(os.Stderr, "", 0)

	format := logging.MustStringFormatter(
		`%{time:15:04:05} %{level:.5s} %{message}`)
	logFormatter := logging.NewBackendFormatter(stderrLog, format)

	logLevel := logging.AddModuleLevel(stderrLog)
	level := logging.ERROR
	if debug {
		level = logging.DEBUG
	}
	logLevel.SetLevel(level, "")

	logging.SetBackend(logFormatter)
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
	initLogging(*debugFlag)

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
	// if err := print_search_results(config); !err {
	//        log.Error(err)
	//        os.exit(1)
	// }

	os.Exit(0)
}
