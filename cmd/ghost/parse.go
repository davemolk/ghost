package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

func (g *ghost) parsePage(page string, query interface{}) {
	seen := make(map[string]bool)

	switch q := query.(type) {
	case *regexp.Regexp:
		results := q.FindAllString(string(page), -1)
		if results == nil {
			fmt.Println("no matches found")
			return
		}
		for _, v := range results {
			if seen[v] {
				continue
			}
			seen[v] = true
			fmt.Println(v)
		}
	case string:
		if strings.Contains(page, q) {
			fmt.Printf("found %s\n", q)
		} else {
			fmt.Printf("failed to find %s\n", q)
		}
	case []string:
		var wg sync.WaitGroup
		for _, term := range q {
			wg.Add(1)
			go func(t string) {
				defer wg.Done()
				if strings.Contains(page, t) {
					fmt.Printf("found %s\n", t)
				} else {
					fmt.Printf("failed to find %s\n", t)
				}
			}(term)
		}
		wg.Wait()
	default:
		log.Fatal("malformed query")
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
	} else {
		g.query = g.config.term
	}
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