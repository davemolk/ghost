package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// getUA returns a string slice of ten user agents.
func (g *ghost) getUA() []string {
	return []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4692.56 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4889.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.84 Safari/537.36",
	}
}

// randomUA returns a randomly selected user agent from the list of 10 found in getUA.
func (g *ghost) randomUA() string {
	userAgents := g.getUA()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rando := r.Intn(len(userAgents))
	return userAgents[rando]
}

// checkAsset checks if the Wayback Machine has a snapshot for a given URL.
// If it does, checkAsset will get the snapshot and write its contents to
// a file.
func (g *ghost) checkAsset(wg *sync.WaitGroup, url, filename string, timeout int) {
	defer wg.Done()

	// call function to parse filename and create URL
	u := g.createURL(url, filename)

	available := g.checkAvailable(u, timeout)
	if available == "" {
		g.errorLog.Printf("unable to get %s\n", u)
		return
	}

	body, err := g.getData(u, timeout)
	if err != nil {
		g.errorLog.Printf("unable to get %s: %v\n", u, err)
		return
	}
	if len(body) > 0 {
		g.writeData(filename, body)
	} else {
		g.errorLog.Printf("no data at %s\n", u)
	}
}

// createURL takes in a URL and a filename and uses the filename to
// determine whether to return URL/robots.txt/ or URL/sitemap.xml/.
func (g *ghost) createURL(url, filename string) string {
	url = strings.TrimSuffix(url, "/")
	var u string
	if strings.HasSuffix(filename, "robots.txt") {
		u = fmt.Sprintf("%s/robots.txt", url)
	} else {
		u = fmt.Sprintf("%s/sitemap.xml", url)
	}
	return u
}

// wayback struct for storing the information coming back
// as json from the Wayback Machine API availability endpoint.
type wayback struct {
	ArchivedSnapshots struct {
		Closest struct {
			Available bool   `json:"available"`
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
			Status    string `json:"status"`
		} `json:"closest"`
	} `json:"archived_snapshots"`
}

// checkAvailable takes in a url and a timeout and checks the
// Wayback Machine API availability endpoint. If the url is not
// available, an empty string is returned. Otherwise, checkAvailable
// returns the URL containing the latest snapshot for the submitted site.
func (g *ghost) checkAvailable(url string, timeout int) string {
	const prefix = "http://archive.org/wayback/available?url="
	u := fmt.Sprintf("%s%s", prefix, url)
	g.infoLog.Printf("checking: %s", u)

	body, err := g.getData(u, timeout)
	if err != nil {
		g.errorLog.Printf("unable to get %s: %v\n", u, err)
		return ""
	}

	wayback := &wayback{}
	err = json.Unmarshal(body, &wayback)
	if err != nil {
		g.errorLog.Printf("%s unmarshal error: %v\n", u, err)
		return ""
	}
	if !wayback.ArchivedSnapshots.Closest.Available {
		g.infoLog.Printf("%s not available\n", u)
		return ""
	} else {
		return wayback.ArchivedSnapshots.Closest.URL
	}
}

// getData takes in a url and a timeout and returns the response body as
// a slice of bytes.
func (g *ghost) getData(url string, timeout int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	uAgent := g.randomUA()
	req.Header.Set("User-Agent", uAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	return body, nil
}

// getSnaps takes in a byte slice (obtained from the cdx server), unmarshals
// it, and returns the wayback machine snapshots in a slice.
func (g *ghost) getSnaps(data []byte) ([][]string, error) {
	var snaps [][]string
	err := json.Unmarshal(data, &snaps)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	if len(snaps) == 0 {
		return nil, errors.New("no wayback machine snapshots found. If using limit=-1, try limit=-2")
	}

	g.writeData("data/snaps.json", data)

	g.infoLog.Printf("Found %d snapshot(s).", len(snaps[1:]))

	// leave off the key
	return snaps[1:], nil
}

// archivedURLs leverages the Wayback Machine API responsible for populating
// all captured URLs associated with a given URL prefix. The data is written
// to an archivedURLs.json file.
func (g *ghost) archivedURLs(wg *sync.WaitGroup, url string, timeout int) {
	defer wg.Done()
	now := time.Now()
	curr := now.UnixMilli()
	const guts = "&matchType=prefix&collapse=urlkey&output=json&fl=original%2Cmimetype%2Ctimestamp%2Cendtimestamp%2Cgroupcount%2Cuniqcount&filter=!statuscode%3A%5B45%5D..&limit=10000&_="
	u := fmt.Sprintf("https://web.archive.org/web/timemap/json?url=%s%s%d", url, guts, curr)
	body, err := g.getData(u, timeout)
	if err != nil {
		g.errorLog.Printf("archivedURLs unsuccessful: %v", err)
		return
	}
	if len(body) > 0 {
		g.sortData(body)
		g.writeData("data/archivedURLs.json", body)
	} else {
		g.errorLog.Println("no archived links on web.archive.org")
	}
}

// sortData takes in a byte slice, unmarshals it, and creates two subsets
// to reflect whether or not each URL in the data set is unique. The two
// subsets are then printed to a file.
func (g *ghost) sortData(data []byte) {
	g.infoLog.Println("Sorting URLs.")

	unique := [][]string{
		{"original", "mimetype", "timestamp", "endtimestamp", "groupcount", "uniqcount"},
	}
	multiple := [][]string{
		{"original", "mimetype", "timestamp", "endtimestamp", "groupcount", "uniqcount"},
	}

	var s [][]string

	err := json.Unmarshal(data, &s)
	if err != nil {
		g.errorLog.Printf("sortData unmarshal error: %v\n", err)
		return
	}

	// skip the key
	for _, v := range s[1:] {
		// check uniqcount field in snapshot
		if g.isUnique(v[5]) {
			unique = append(unique, v)
		} else {
			multiple = append(multiple, v)
		}
	}
	if len(unique) > 0 {
		b, err := g.JSON(unique)
		if err != nil {
			g.errorLog.Printf("sortData marshal error: %v\n", err)
			return
		}
		g.writeData("data/unique.json", b)
	}
	if len(multiple) > 0 {
		b, err := g.JSON(multiple)
		if err != nil {
			g.errorLog.Printf("sortData marshal error: %v\n", err)
			return
		}
		g.writeData("data/multiple.json", b)
	}
}

// isUnique takes in a string and compares it to "1", returning true
// if they match and false otherwise.
func (g *ghost) isUnique(s string) bool {
	return s == "1"
}

// JSON uses NewEncoder over Marshal in order to avoid the escaped HTML.
func (g *ghost) JSON(data [][]string) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(data)
	return bytes.TrimRight(buf.Bytes(), "\n"), err
}

// whoisLookup checks the "whois.iana.org" server to find information about the domain. Any
// results are then written to "whois.txt" via the writeData function.
func (g *ghost) whoisLookup(wg *sync.WaitGroup, domain string, timeout int) {
	defer wg.Done()
	// incorporate lookup if there eventually becomes a need past iana
	port := "43"
	server := "whois.iana.org"

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	var d = &net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(server, port))
	if err != nil {
		g.errorLog.Printf("whois connection failure: %v\n", err)
		return
	}

	defer conn.Close()

	_, err = conn.Write([]byte(domain + "\r\n"))
	if err != nil {
		g.errorLog.Printf("send to whois failure: %v\n", err)
		return
	}

	buff, err := io.ReadAll(conn)
	if err != nil {
		g.errorLog.Printf("whois read failure: %v\n", err)
		return
	}

	if len(buff) > 0 {
		g.writeData("data/whois.txt", buff)
	} else {
		g.infoLog.Println("No results for whois.")
	}
}

// getIP takes in a host and writes the IPv4 and IPv6 addresses
// to a file.
func (g *ghost) getIP(wg *sync.WaitGroup, host string) {
	defer wg.Done()
	ips, err := net.LookupIP(host)
	if err != nil {
		g.errorLog.Println("unable to look up IP")
		return
	}

	var ipByte []byte
	for _, ip := range ips {
		b, err := ip.MarshalText()
		if err != nil {
			g.errorLog.Println("marshal error in getIP")
			return
		}
		ipByte = append(ipByte, b...)
		// add breaks
		ipByte = append(ipByte, byte(0x0A))
	}

	g.writeData("data/ip.txt", ipByte)
}
