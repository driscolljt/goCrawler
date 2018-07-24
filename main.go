package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// Basic scraper that takes in a url and outputs all unique urls found in the response html
func main() {
	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]

	// channels
	chUrls := make(chan string)
	chFinished := make(chan bool)

	// kick off the crawl process (concurrently)
	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	// we're done, print the results...

	fmt.Println("\nFound", len(foundUrls), "unique urls:")

	for url, _ := range foundUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}

// pulls the href attribute out of an anchor tag
func getHref(t html.Token) (ok bool, href string) {
	// find the href
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	// calling a 'bare' return will pull
	// the attribute names in the func definition
	// 'ok' and 'href'
	return
}

// Extract all links from a given web page
func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		// notify that we are done on the 'chFinished' channel
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close body when function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// end of document
			return

		case tt == html.StartTagToken:
			t := z.Token()

			// check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if isAnchor {
				continue
			}

			// extract the href value, if there is one
			ok, url := getHref(t)
			if !ok {
				continue
			}

			// make sure the url beings with http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}
