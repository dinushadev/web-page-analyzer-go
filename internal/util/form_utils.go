package util

import (
	"strings"

	"golang.org/x/net/html"
)

func HasLoginForm(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "form" {
		if ContainsPasswordInput(n) {
			return true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if HasLoginForm(c) {
			return true
		}
	}
	return false
}

func ContainsPasswordInput(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "input" {
		for _, attr := range n.Attr {
			if strings.EqualFold(attr.Key, "type") && strings.EqualFold(attr.Val, "password") {
				return true
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if ContainsPasswordInput(c) {
			return true
		}
	}
	return false
}


