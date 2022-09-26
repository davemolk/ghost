package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// writeJSON takes in a byte slice and file name and writes
// the contents to a .txt file.
func (g *ghost) writeJSON(name string, data []byte) {
	g.infoLog.Printf("writing %s", name)
	f, err := os.Create(name)
	if err != nil {
		g.errorLog.Println(err)
		return
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		g.errorLog.Println(err)
		return
	}
	err = f.Sync()
	if err != nil {
		g.errorLog.Println(err)
	}
}

// searchMapWriter determines a file name based on the query type, marshals
// the searchMap, and then calls writeJSON.
func (g *ghost) searchMapWriter(query interface{}, data map[string][]string) {
	var name string
	switch query.(type) {
	case string:
		name = "termResults.json"
	case []string:
		name = "termsResults.json"
	case *regexp.Regexp:
		name = "regexResults.json"
	}

	b, err := json.Marshal(data)
	if err != nil {
		g.errorLog.Printf("Marshal error: %v\n", err)
		return
	}

	g.writeJSON(name, b)

}

// getQuery checks whether the user has submitted a search term flag, a
// regex flag, or a file input flag and creates the query accordingly.
func (g *ghost) getQuery() bool {
	switch {
	case len(g.config.regex) > 0:
		g.query = regexp.MustCompile(g.config.regex)
		return true
	case len(g.config.terms) > 0:
		query, err := g.readInputFile(g.config.terms)
		if err != nil {
			g.errorLog.Fatal("Unable to read input file")
		}
		g.query = query
		return true
	case len(g.config.term) > 0:
		g.query = g.config.term
		return true
	default:
		g.errorLog.Println("No query submitted.")
		return false
	}
}

// formURL takes in the query parameters and forms the search URL for the
// CDX server. Including default values of "" doesn't impact the query results.
func (g *ghost) formURL(url, mimetype, from, to string, limit, statuscode int) string {
	const base = "http://web.archive.org/cdx/search/cdx?output=json"
	u := fmt.Sprintf("%s&fastLatest=true&url=%s&mimetype=%s&from=%s&to=%s&limit=%d&filter=statuscode:%d", base, url, mimetype, from, to, limit, statuscode)
	return u
}

// readInputFile reads and converts the contents of an input text file
// to a string slice, returning that and any errors.
func (g *ghost) readInputFile(name string) ([]string, error) {
	var lines []string

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines, s.Err()
}

// getInputURL accepts a URL from stdin and sets it to g.config.url.
func (g *ghost) getInputURL() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		g.config.url = s.Text()
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "unable to read input: %v", err)
	}
	if g.config.url == "" {
		g.errorLog.Fatal("missing input url")
	}
}