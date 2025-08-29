package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

func hasLoginForm(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "form" {
		if containsPasswordInput(n) {
			logInfo("login_form.detected")
			return true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasLoginForm(c) {
			return true
		}
	}
	return false
}

func containsPasswordInput(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "input" {
		for _, attr := range n.Attr {
			if strings.EqualFold(attr.Key, "type") && strings.EqualFold(attr.Val, "password") {
				return true
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if containsPasswordInput(c) {
			return true
		}
	}
	return false
}
