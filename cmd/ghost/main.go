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
	domain        string
	from          string
	host          string
	limit         string
	mimetype      string
	notMimetype   string
	notStatusCode string
	prefix        string
	statuscode    string
	to            string
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
	flag.IntVar(&config.gophers, "g", 10, "number of goroutines (default is 10).")
	flag.StringVar(&config.regex, "regex", "", "regex pattern for parsing search results.")
	flag.StringVar(&config.term, "term", "", "term for parsing search results.")
	flag.StringVar(&config.terms, "terms", "", "name of file containing term list for parsing search results.")
	flag.IntVar(&config.timeout, "time", 5000, "timeout in milliseconds (default is 5000).")
	flag.StringVar(&config.url, "u", "", "url for searching")

	// filtering Wayback Machine results
	flag.StringVar(&config.filters.from, "f", "", "search from here, including at least a year. format more specific queries as yyyyMMddhhmmss.")
	flag.StringVar(&config.filters.limit, "l", "0", "limit query results, using -1, -2, -3 etc. for most recent, 1, 2, 3 etc. for oldest.")
	flag.StringVar(&config.filters.mimetype, "m", "text/html", "filter results according to mimetype (default is 'text/html').")
	flag.StringVar(&config.filters.notMimetype, "nm", "", "filter specified mimetype out of results (inactive by default).")
	flag.StringVar(&config.filters.notStatusCode, "ns", "0", "filter specified status code out of results (inactive by default).")
	flag.StringVar(&config.filters.statuscode, "s", "200", "filter results by status code (default is 200).")
	flag.StringVar(&config.filters.to, "t", "", "search to here, including at least a year. format more specific queries as yyyyMMddhhmmss.")

	// matchType
	flag.StringVar(&config.filters.domain, "domain", "", "return results from host and all subhosts.")
	flag.StringVar(&config.filters.host, "host", "", "return results from host.")
	flag.StringVar(&config.filters.prefix, "prefix", "", "return results for all results under the path.")

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

	if config.url != "" {
		g.validateURL(config.url)
	} else {
		g.getInputURL()
	}

	var wg sync.WaitGroup

	host, err := g.getHost(g.config.url)
	if err != nil {
		g.errorLog.Printf("getHost error: %v\n", err)
	}
	wg.Add(1)
	go g.getIP(&wg, host)

	domain, err := g.getDomain(g.config.url)
	if err != nil {
		g.errorLog.Printf("getDomain error: %v\n", err)
	}
	if domain != "" { // NEED THIS?
		wg.Add(1)
		go g.whoisLookup(&wg, domain, config.timeout)
	}

	validQuery := g.getQuery()
	u := g.formURL(g.config.url, config.filters)
	g.infoLog.Printf("Wayback Machine URL: %s\n", u)

	g.client = g.makeClient(config.timeout)

	// get all captured resources for given URL prefix
	wg.Add(1)
	go g.getResources(&wg, g.client, g.config.url)

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

	// wait here in case of early exit cause no query
	wg.Wait()
	
	if !validQuery {
		g.infoLog.Print("Snapshots retrieved and saved to file. Exiting...")
		os.Exit(1)
	}

	// extract timestamps from snaps
	var filteredSnaps []string
	for _, v := range snaps {
		filteredSnaps = append(filteredSnaps, v[1])
	}

	tokens := make(chan struct{}, config.gophers)

	for _, timestamp := range filteredSnaps {
		tokens <- struct{}{}
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			url := fmt.Sprintf("https://web.archive.org/web/%s/%s", t, g.config.url)
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

	fmt.Printf("Took: %f seconds\n", time.Since(start).Seconds())
}
