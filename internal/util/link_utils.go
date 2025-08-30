package util

import (
	"net/url"

	"golang.org/x/net/html"
)

func CountLinks(n *html.Node, base *url.URL, isAccessible func(string) bool) (internal, external, inaccessible int) {
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
				if !isAccessible(abs.String()) {
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
