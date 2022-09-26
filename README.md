# ghost
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)
// // FINISH go report card here
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/davemolk/ghost/issues)

Get and parse a URL's Wayback Machine history with ghost.

## Overview
* Supply a URL and get a file containing all archived snapshots and another containing all archived links for that URL prefix. Add a query to search each snapshot. Search results are written to a additional file.
* Use -term, -regex, and -terms, respectively, to scan each page for a specific word, a list of words (input as a .txt file), or with a regular expression.
* Customize your search with advanced query filtering and URL matching based on different parameters.
* See [here](https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server) for more details on the Wayback cdx server.

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
  -l string
    	Limit query results, using -1, -2, -3 etc. for most recent and 1, 2, 3 etc. for oldest.
  -m string
        Filter results according to mimetype (default is 'text/html').
  -nm string
        Filter specified mimetype out of results (inactive by default).
  -ns string
        Filter specified status code out of results (inactive by default).
  -s string
        Filter results by status code (default is 200).
  -t string
    	Search to here, including at least a year. Format more specific queries as yyyyMMddhhmmss.

(match scope)
  -domain string
        Return results from host and all subhosts (inactive by default).
  -host string
        Return results from host (inactive by default).
  -prefx string
        Return results for all results under the path (inactive by default).

```

## Installation
First, you'll need to [install go](https://golang.org/doc/install).

Then run this command to download + compile ghost:
```
go install github.com/davemolk/ghost/cmd/ghost@latest
```
Alternatively, use one of the binaries available in the release.

## Additional Notes
* Occasionally, a limit of -1 erroneously returns no results (this happens with curl and in a browser). If this happens and you know you should be seeing something, use limit of -2.
* The query string in formURL contains "fastLatest=true." I haven't noticed an appreciable difference, but it can't hurt, right? See more [here](https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server):
* The query string also contains &collapse=digest, collapsing adjacent digests for less cluttered results.

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