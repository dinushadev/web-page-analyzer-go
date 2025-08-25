package service

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"test-project-go/internal/model"
)

var ErrUnreachable = errors.New("URL is unreachable")

func AnalyzePage(targetURL string) (*model.AnalyzeResult, int, error) {
	parsed, err := url.ParseRequestURI(targetURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, 0, errors.New("invalid URL")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, 0, ErrUnreachable
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, resp.StatusCode, errors.New("HTTP error: " + resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, 0, errors.New("failed to parse HTML")
	}

	result := &model.AnalyzeResult{}
	result.HTMLVersion = detectHTMLVersion(doc)
	result.Title = extractTitle(doc)
	result.Headings = countHeadings(doc)
	internal, external, inaccessible := countLinks(doc, parsed)
	result.Links = model.LinkStats{Internal: internal, External: external, Inaccessible: inaccessible}
	result.LoginForm = hasLoginForm(doc)
	return result, 0, nil
}

func detectHTMLVersion(n *html.Node) string {
	if n.Type == html.DoctypeNode {
		dt := strings.ToLower(n.Data)
		if strings.Contains(dt, "html") {
			return "HTML5"
		}
		return n.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if v := detectHTMLVersion(c); v != "" {
			return v
		}
	}
	return "Unknown"
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
		return n.FirstChild.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := extractTitle(c); t != "" {
			return t
		}
	}
	return ""
}

func countHeadings(n *html.Node) []model.HeadingCount {
	counts := make(map[int]int)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && len(n.Data) == 2 && n.Data[0] == 'h' && n.Data[1] >= '1' && n.Data[1] <= '6' {
			level := int(n.Data[1] - '0')
			counts[level]++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	var result []model.HeadingCount
	for i := 1; i <= 6; i++ {
		if counts[i] > 0 {
			result = append(result, model.HeadingCount{Level: i, Count: counts[i]})
		}
	}
	return result
}

func countLinks(n *html.Node, base *url.URL) (internal, external, inaccessible int) {
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	for _, l := range links {
		u, err := url.Parse(l)
		if err != nil || u.Scheme == "javascript" || u.Scheme == "mailto" {
			continue
		}
		abs := u
		if !u.IsAbs() {
			abs = base.ResolveReference(u)
		}
		if abs.Host == base.Host {
			internal++
		} else {
			external++
		}
		if !isLinkAccessible(abs.String()) {
			inaccessible++
		}
	}
	return
}

func isLinkAccessible(link string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Head(link)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func hasLoginForm(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "form" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "input" {
				for _, attr := range c.Attr {
					if attr.Key == "type" && attr.Val == "password" {
						return true
					}
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasLoginForm(c) {
			return true
		}
	}
	return false
}
