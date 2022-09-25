# ghost
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)
// // FINISH go report card here
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/davemolk/ghost/issues)

Get and parse a URL's Wayback Machine history with ghost.

## Overview
* // FINISH
* Use -term, -regex, and -terms, respectively, to scan each page for a specific word, with a regular expression, or with a list of words (input as a .txt file).
* Search results are written to a .json file.
* ghost also collects every URL archived by Wayback Machine (within the specified date range) and writes them to a .json file.

## Example Usage
```
go run ./cmd/ghost -u https://go.dev -f 20220922 -time 10000 -term go -l -2
```

## Command-line Options
```
Usage of ghost:
  -g int
    	Number of goroutines to use (default is 10).
  -r string
    	Regex pattern for parsing search results.
  -term string
    	Term for parsing search results.
  -terms string
    	Name of a .txt file containing a list of terms for parsing search results.
  -time int
    	Request timeout (in milliseconds). Default is 5000.
  -u string
    	URL for searching.

(query filtering)
  -f string
    	Search from here, including at least a year. Format more specific queries as yyyyMMddhhmmss.
  -l int
    	Limit query results, using -1, -2, -3 etc. for most recent and 1, 2, 3 etc. for oldest.
  -s int
        Filter results by status code (default is 200).
  -t string
    	Search to here, including at least a year. Format more specific queries as yyyyMMddhhmmss.
```

## Installation
First, you'll need to [install go](https://golang.org/doc/install).

Then run this command to download + compile ghost:
```
go install github.com/davemolk/ghost/cmd/ghost@latest
```

## Additional Notes
* Occasionally, a limit of -1 erroneously returns no results (this happens with curl and in a browser). If this happens and you know you should be seeing something, use limit of -2.
* The query string in formURL contains "fastLatest=true." I haven't noticed an appreciable difference, but it can't hurt, right? See more [here](https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server):

## Changelog
*    **2022-09-26** : ghost

## Support
* Like ghost? Use it, star it, and share with your friends!
    - Let me know what you're up to so I can feature your work here.
* Want to see a particular feature? Found a bug? Question about usage or documentation?
    - Please raise an issue.
* Pull request?
    - Please discuss in an issue first. 

## License
* ghost is released under the MIT license. See [LICENSE](LICENSE) for details.