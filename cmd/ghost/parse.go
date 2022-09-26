package main

import (
	"regexp"
	"strings"
	"sync"
)

// parsePage takes in a page and searches its contents for whatever
// query the user submitted (regular expression, a single search term,
// or a list of terms supplied in a .txt file).
func (g *ghost) parsePage(page, url string, query interface{}) {
	seen := make(map[string]bool)

	switch q := query.(type) {
	case *regexp.Regexp:
		results := q.FindAllString(string(page), -1)
		if results == nil {
			g.infoLog.Printf("Failed to find %v.\n", q)
			return
		}
		for _, result := range results {
			if seen[result] {
				continue
			}
			seen[result] = true
			g.searches.store(result, url)
		}
	case string:
		if len(q) > 0 && strings.Contains(page, q) {
			g.searches.store(q, url)
		} else {
			g.infoLog.Printf("Failed to find %s.\n", q)
		}
	case []string:
		var wg sync.WaitGroup
		for _, term := range q {
			wg.Add(1)
			go func(t string) {
				defer wg.Done()
				if strings.Contains(page, t) {
					g.searches.store(t, url)
				} else {
					g.infoLog.Printf("Failed to find %s.\n", t)
				}
			}(term)
		}
		wg.Wait()
	}
}

// searchMap is a mutex-protected map that stores the search results
// in the key-value form query: url(s).
type searchMap struct {
	mu       sync.Mutex
	searches map[string][]string
}

// newSearchMap returns a pointer to a new searchMap.
func newSearchMap() *searchMap {
	return &searchMap{
		searches: make(map[string][]string),
	}
}

// store takes in a term and a url where it was found, locks
// the searchMap, stores the information, and unlocks the searchMap.
func (s *searchMap) store(term, url string) {
	s.mu.Lock()
	s.searches[term] = append(s.searches[term], url)
	s.mu.Unlock()
}
