# ghost
Get and parse a URL's Wayback Machine history. Save all archived links for that URL prefix while you're at it.

![demo](ghost.gif)

## Overview
* Supply a URL and get a file containing all archived snapshots and another containing all archived links for that URL prefix. Use -term, -terms, or -regex, to scan each snapshot for a specific word, a list of words (input as a .txt file), or with a regular expression. All search results are saved to a file.
* Customize your search with advanced query filtering.
* In addition to exact URL matching (default), ghost also supports URL matching based on -domain, -host, and -prefix.

## Example Usage
(find the two most recent results from https://go.dev, starting at 9/22/2022 and using a 10-second timeout.)
```
go run ./cmd/ghost -u https://go.dev -f 20220922 -time 10000 -term go -l -2
```
(download a binary or run go build ./cmd/ghost, then run)
```
echo https://go.dev | ./ghost -f 20220922 -time 10000 -term go -l -2
```
(install with go install ./cmd/ghost, then run)
```
echo https://go.dev | ghost -f 20220922 -time 10000 -term go -l -2
```
## Command-line Options
```
Usage of ghost:
  -g int
    	Number of goroutines (default is 10).
  -regex string
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
  -prefix string
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
* Occasionally, a limit of -1 erroneously returns no results (this also happens when using curl or a browser). If you know you should be seeing something and this happens, use limit of -2.
* The query string in formURL contains "fastLatest=true." I haven't noticed an appreciable difference, but it can't hurt, right? Visit [here](https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server) for more details.
* The query string also contains &collapse=digest, which collapses adjacent digests for less cluttered results.

## Changelog
*    **2022-09-27** : ghost v1.0

## Support
* Like ghost? Use it, star it, and share with your friends!
* Want to see a particular feature? Found a bug? Question about usage or documentation?
    - Great! Let me know.
* Pull request?
    - Please discuss in an issue first. 

## License
* ghost is released under the MIT license. See [LICENSE](LICENSE) for details.



#### ...the latch was left unhooked...