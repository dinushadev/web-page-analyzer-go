package analyzer

import (
	"net/url"

	"golang.org/x/net/html"
)

func countLinks(n *html.Node, base *url.URL, checker LinkChecker) (internal, external, inaccessible int) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			link := ""
			for _, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					link = attr.Val
					break
				}
			}
			if link != "" {
				u, err := url.Parse(link)
				if err != nil || u.Scheme == "javascript" || u.Scheme == "mailto" {
					return
				}
				abs := u
				if !u.IsAbs() {
					abs = base.ResolveReference(u)
				}
				if !checker.IsAccessible(abs.String()) {
					inaccessible++
					return
				}
				if abs.Host == base.Host {
					internal++
				} else {
					external++
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return
}
