package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

// writeJSON takes in a byte slice and file name and writes
// the contents to a .txt file.
func (g *ghost) writeJSON(name string, data []byte) {
	f, err := os.Create(name)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Println(err)
		return
	}
	err = f.Sync()
	if err != nil {
		log.Println(err)
	}
}

// getQuery checks whether the user has submitted a search term flag, a
// regex flag, or a file input flag and creates the query accordingly.
func (g *ghost) getQuery() {
	if len(g.config.regex) > 0 {
		g.query = regexp.MustCompile(g.config.regex)
	} else if len(g.config.terms) > 0 {
		query, err := g.readInputFile(g.config.terms)
		if err != nil {
			log.Fatal("unable to read input file")
		}
		g.query = query
	} else if len(g.config.term) > 0 {
		g.query = g.config.term
	} else {
		log.Println("note: no query supplied")
	}
}

// formURL takes in the query parameters and forms the search URL for the
// CDX server. Including default values of "" doesn't impact the query results.
func (g *ghost) formURL(url, from, to string, limit, statuscode int) string {
	const base = "http://web.archive.org/cdx/search/cdx?output=json"
	u := fmt.Sprintf("%s&fastLatest=true&url=%s&from=%s&to=%s&limit=%d&filter=statuscode:%d", base, url, from, to, limit, statuscode)
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
