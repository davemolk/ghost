package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
)

// parsePage takes in a page and searches its contents for whatever
// query the user submitted (regular expression, a single search term,
// or a list of terms supplied in a .txt file).
func (g *ghost) parsePage(page string, query interface{}) {
	seen := make(map[string]bool)

	switch q := query.(type) {
	case *regexp.Regexp:
		results := q.FindAllString(string(page), -1)
		if results == nil {
			fmt.Printf("failed to find %v\n", q)
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
		if len(q) > 0 && strings.Contains(page, q) {
			fmt.Printf("found %s\n", q)
		} else if len(q) > 0 {
			fmt.Printf("failed to find %s\n", q)
		} else {
			log.Println("no search query specified")
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
	}
}
