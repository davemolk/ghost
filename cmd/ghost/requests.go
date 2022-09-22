package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// makeClient takes in a timeout (in milliseconds) and returns a client.
func (g *ghost) makeClient(timeout int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
}

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

// makeRequest takes in a url and a client and returns an http.Response and an error.
func (g *ghost) makeRequest(url string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	uAgent := g.randomUA()
	req.Header.Set("User-Agent", uAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return resp, nil
}

// getData takes in a url and a client and returns the response body as
// a slice of bytes.
func (g *ghost) getData(url string, client *http.Client) ([]byte, error) {
	resp, err := g.makeRequest(url, client)
	if err != nil {
		return nil, fmt.Errorf("makeRequest error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	return body, nil
}

// getSnaps takes in a byte slice and unmarshals it, returning the 
// snapshots from the wayback machine in a slice.
func (g *ghost) getSnaps(data []byte) ([][]string, error) {
	var snaps [][]string
	err := json.Unmarshal(data, &snaps)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	if len(snaps) == 0 {
		return nil, errors.New("no snapshots available from wayback machine")
	}
	return snaps[1:], nil
}