
*WARNING -- this is a work in progress; the code is not complete*

# booklist

This is the equivalent of the GitHub repository for 'booklist', but written
in Go.

If you're a frequent library user, you might go online to your library's
public website and search for the latest books available from your favorite
authors.  If your list of favorite authors is long, that's a tedious task
even with saved searches.

`booklist` can make that task easier by automating the task of
determining whether the library has any new publications for those authors.
However, this tool will only work with libraries using the CARL.X
Integrated Library System.

`booklist` works by accessing a given library's website to search for
publications from a given author, of a specific media type and within the
current year.  The results are printed to the console.  The search is
repeated for each author listed in a configuration file.

The configuration file is in YAML format and provides the library's
catalog URL, default media type, authors and any specific media types
for authors.  The publication time period includes publications with an
unknown publication time as the unknown time might include the current year.

## Installation

To install the latest release from github:

```sh
go get github.com/kbalk/go_booklist
```

After installation, modify the default configuration file `sample_config.yml`,
which is found in the same directory as this README.  In particular, the
catlog-url and author list should be edited to specify the appropriate
library and authors of interest.  Refer to the
['Configuration File'](#configuration-file) section for further details.

### Prerequisites

This tool was written and tested using Go 1.8.  It will most likely work
with earlier versions of Go.

## Configuration File

The name of the YAML-formatted configuration file is a required argument
to run `booklist`.  The format is as follows:

Tag   | Description
------------------|-----------------
catalog-url | Required.  Must be a valid URL for a website using the CARL.X Integrated Library System.
media-type | Optional.  The default media type is book; allowed types are listed below.
authors     | Required.  List of authors specified by first and last name and optionally by media-type.
firstname   | Required.  Sub-tag of 'authors'.  First name of author.
lastname    | Required.  Sub-tag of 'authors'.  Last name of author.
media-type | Optional.  Sub-tag of 'authors'.  See list of media types below.

Allowed media types:

- Book
- Electronic Resource
- eBook
- eAudioBook
- Book on CD
- Large Print
- Music CD
- DVD
- Blu-Ray

Note:  The media types can be in upper, lower or mixed case as the case
will be ignored.  Types that consist of more than one word may be enclosed
in quotes or not.

Also, some media types are supersets, i.e., a type of 'book' includes
'large print' books.  A type of 'electronic resource' includes 'ebook'.

Example configuration file:

```YAML
catalog-url: http://catalog.library.loudoun.gov/
media-type: Book
authors:
   - firstname: James
     lastname: Patterson
     media-type: book on cd

   - firstname: Alexander
     lastname: McCall Smith
```

## Usage

```sh
Usage: booklist [-h] [-d] config_file

Search a public library's catalog website for this year's publications
from authors listed in the given config file.

positional arguments:
  config_file  YAML formatted file containing library's catalog url and
               list of authors

optional arguments:
  -h  show this help message and exit
  -d  Print debug information to stdout
```

A sample configuration file named `sample_config.yml` has been provided with
the distribution.  The format of the configuration file is described
[here](#configuration-file).

## Limitations

I wasn't able to determine the version of CARL.X used in my testing,
but even so it appears that CARL.X is configurable for the types of
filters and media it will permit.  Therefore this tool may not work
for any given CARL.X ILS, but it can be used as a starting point
for any needed modifications.
