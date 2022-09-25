package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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
	errorLog *log.Logger
	infoLog  *log.Logger
	query    interface{}
	searches *searchMap
}

func main() {
	var config config
	flag.IntVar(&config.gophers, "g", 10, "number of goroutines to use (default is 10)")
	flag.StringVar(&config.regex, "r", "", "regex pattern for parsing search results")
	flag.StringVar(&config.term, "term", "", "term or phrase for parsing search results")
	flag.StringVar(&config.terms, "terms", "", "name of file containing term list for parsing search results")
	flag.IntVar(&config.timeout, "time", 5000, "timeout in milliseconds (default is 5000)")
	flag.StringVar(&config.url, "u", "", "url for searching")

	flag.StringVar(&config.filters.from, "f", "", "search from here, including at least a year. format more specific queries as yyyyMMddhhmmss")
	flag.IntVar(&config.filters.limit, "l", 0, "limit query results, using -1, -2, -3 etc. for most recent, 1, 2, 3 etc. for oldest")
	flag.IntVar(&config.filters.statuscode, "s", 200, "filter results by status code (default is 200)")
	flag.StringVar(&config.filters.to, "t", "", "search to here, including at least a year. format more specific queries as yyyyMMddhhmmss")

	flag.Parse()

	start := time.Now()

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	searches := newSearchMap()

	g := &ghost{
		config:   config,
		errorLog: errorLog,
		infoLog:  infoLog,
		searches: searches,
	}

	if config.url == "" {
		g.errorLog.Fatal("URL must be provided")
	}

	validQuery := g.getQuery()
	u := g.formURL(config.url, config.filters.from, config.filters.to, config.filters.limit, config.filters.statuscode)

	g.client = g.makeClient(config.timeout)

	// get all captured resources for URL prefix
	done := make(chan bool)
	go g.getResources(g.client, config.url, done)

	// check Wayback Machine
	body, err := g.getData(u, g.client)
	if err != nil {
		g.errorLog.Fatal(err)
	}

	// also saves the snaps to a .json file
	snaps, err := g.getSnaps(body)
	if err != nil {
		g.errorLog.Fatal(err)
	}

	if !validQuery {
		g.infoLog.Fatal("Snapshots retrieved and saved to file. Exiting...")
	}

	// extract timestamps from snaps
	var filteredSnaps []string
	for _, v := range snaps {
		filteredSnaps = append(filteredSnaps, v[1])
	}

	var wg sync.WaitGroup
	tokens := make(chan struct{}, config.gophers)

	for _, timestamp := range filteredSnaps {
		tokens <- struct{}{}
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			url := fmt.Sprintf("https://web.archive.org/web/%s/%s", t, config.url)
			page, err := g.getData(url, g.client)
			if err != nil {
				g.errorLog.Printf("getData error for %s: %v\n", url, err)
				<-tokens
				return
			}
			<-tokens
			g.parsePage(string(page), url, g.query)
		}(timestamp)
	}
	wg.Wait()

	g.searchMapWriter(g.query, g.searches.searches)

	// make sure getResources has finished
	<- done

	fmt.Printf("took: %f seconds\n", time.Since(start).Seconds())
}
