package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

type config struct {
	filters filters
	gophers int
	regex   string
	term    string
	terms   string
	timeout int
	url     string
}

type filters struct {
	from       string
	limit      int
	statuscode int
	to         string
}

type ghost struct {
	client   *http.Client
	config   config
	query    interface{}
	searches *searchMap
}

func main() {
	var config config
	flag.IntVar(&config.gophers, "g", 10, "number of goroutines to run")
	flag.StringVar(&config.regex, "r", "", "regex pattern for searching")
	flag.StringVar(&config.term, "term", "", "term or phrase for searching")
	flag.StringVar(&config.terms, "terms", "", "file containing term list for searching")
	flag.IntVar(&config.timeout, "time", 0, "timeout in milliseconds")
	flag.StringVar(&config.url, "u", "", "url for searching")

	flag.StringVar(&config.filters.from, "f", "", "include at least a year. for more specific queries, use format: yyyyMMddhhmmss")
	flag.IntVar(&config.filters.limit, "l", 0, "-1 for most recent, 1 for oldest")
	flag.IntVar(&config.filters.statuscode, "s", 200, "filter results by status code")
	flag.StringVar(&config.filters.to, "t", "", "include at least a year. for more specific queries, use format: yyyyMMddhhmmss")

	flag.Parse()

	start := time.Now()

	searches := newSearchMap()

	g := &ghost{
		config:   config,
		searches: searches,
	}

	g.client = g.makeClient(config.timeout)
	g.query = regexp.MustCompile(config.regex)

	tokens := make(chan struct{}, config.gophers)

	if config.url == "" {
		log.Fatal("search url must be provided")
	}

	g.getQuery()
	u := g.formURL(config.url, config.filters.from, config.filters.to, config.filters.limit, config.filters.statuscode)

	body, err := g.getData(u, g.client)
	if err != nil {
		log.Fatal(err)
	}

	snaps, err := g.getSnaps(body)
	if err != nil {
		fmt.Println(err)
	}

	// extract timestamps from snaps
	var filteredSnaps []string
	for _, v := range snaps {
		filteredSnaps = append(filteredSnaps, v[1])
	}

	var wg sync.WaitGroup
	for _, u := range filteredSnaps {
		tokens <- struct{}{}
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			url = fmt.Sprintf("https://web.archive.org/web/%s/%s", url, config.url)
			page, err := g.getData(url, g.client)
			if err != nil {
				log.Printf("error within getData for %s: %v\n", url, err)
				<-tokens
				return
			}
			<-tokens
			g.parsePage(string(page), url, g.query)
		}(u)
	}

	wg.Wait()

	fmt.Println(g.searches.searches)
	fmt.Printf("took: %f seconds\n", time.Since(start).Seconds())
}
