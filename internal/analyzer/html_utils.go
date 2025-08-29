package analyzer

import (
	"strings"
	"test-project-go/internal/model"

	"golang.org/x/net/html"
)

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
		result = append(result, model.HeadingCount{Level: i, Count: counts[i]})
	}
	return result
}
