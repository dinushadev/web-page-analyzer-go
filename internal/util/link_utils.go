package util

import (
	"net/url"
	"sync"

	"golang.org/x/net/html"
)

func CountLinks(n *html.Node, base *url.URL, isAccessible func(string) bool) (internal, external, inaccessible int) {
	// First pass: collect absolute URLs and track occurrences + internal/external classification
	type linkAgg struct {
		isInternal  bool
		occurrences int
	}
	links := make(map[string]*linkAgg)

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			link := ""
			for _, attr := range node.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					link = attr.Val
					break
				}
			}
			if link != "" {
				u, err := url.Parse(link)
				if err == nil && u.Scheme != "javascript" && u.Scheme != "mailto" {
					abs := u
					if !u.IsAbs() {
						abs = base.ResolveReference(u)
					}
					if abs.Scheme == "http" || abs.Scheme == "https" {
						key := abs.String()
						agg, ok := links[key]
						if !ok {
							agg = &linkAgg{isInternal: abs.Host == base.Host}
							links[key] = agg
						}
						agg.occurrences++
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if len(links) == 0 {
		return 0, 0, 0
	}

	// Second pass: check accessibility concurrently for unique URLs
	// Bounded worker pool to avoid unbounded concurrency
	workerCount := 20
	if workerCount > len(links) {
		workerCount = len(links)
	}
	jobs := make(chan string, workerCount)
	results := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for u := range jobs {
			ok := isAccessible(u)
			mu.Lock()
			results[u] = ok
			mu.Unlock()
		}
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}

	for u := range links {
		jobs <- u
	}
	close(jobs)
	wg.Wait()

	// Aggregate counts with accessibility results and original occurrences
	for u, agg := range links {
		accessibleOK := results[u]
		if !accessibleOK {
			inaccessible += agg.occurrences
			continue
		}
		if agg.isInternal {
			internal += agg.occurrences
		} else {
			external += agg.occurrences
		}
	}
	return
}
