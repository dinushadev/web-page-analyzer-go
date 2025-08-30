package analyzer

import (
	"strconv"
	"strings"
	"test-project-go/internal/model"

	"golang.org/x/net/html"
)

func detectHTMLVersion(n *html.Node) string {
	// Find the first Doctype node
	var findDoctype func(*html.Node) *html.Node
	findDoctype = func(node *html.Node) *html.Node {
		if node == nil {
			return nil
		}
		if node.Type == html.DoctypeNode {
			return node
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if dt := findDoctype(c); dt != nil {
				return dt
			}
		}
		return nil
	}

	dt := findDoctype(n)
	if dt == nil {
		return "Unknown"
	}

	name := strings.ToLower(strings.TrimSpace(dt.Data))
	var publicID, systemID string
	for _, a := range dt.Attr {
		switch strings.ToLower(a.Key) {
		case "public":
			publicID = strings.TrimSpace(a.Val)
		case "system":
			systemID = strings.TrimSpace(a.Val)
		}
	}

	// HTML5: <!DOCTYPE html> with no PUBLIC/SYSTEM identifiers
	if name == "html" && publicID == "" && systemID == "" {
		return "HTML5"
	}

	// Map known PUBLIC identifiers to specific versions
	pid := strings.ToUpper(publicID)
	switch {
	// XHTML 1.x
	case strings.Contains(pid, "XHTML 1.1"):
		return "XHTML 1.1"
	case strings.Contains(pid, "XHTML 1.0") && strings.Contains(pid, "STRICT"):
		return "XHTML 1.0 Strict"
	case strings.Contains(pid, "XHTML 1.0") && strings.Contains(pid, "TRANSITIONAL"):
		return "XHTML 1.0 Transitional"
	case strings.Contains(pid, "XHTML 1.0") && strings.Contains(pid, "FRAMESET"):
		return "XHTML 1.0 Frameset"

	// HTML 4.01
	case strings.Contains(pid, "HTML 4.01") && strings.Contains(pid, "STRICT"):
		return "HTML 4.01 Strict"
	case strings.Contains(pid, "HTML 4.01") && strings.Contains(pid, "TRANSITIONAL"):
		return "HTML 4.01 Transitional"
	case strings.Contains(pid, "HTML 4.01") && strings.Contains(pid, "FRAMESET"):
		return "HTML 4.01 Frameset"

	// Older HTML
	case strings.Contains(pid, "HTML 3.2"):
		return "HTML 3.2"
	case strings.Contains(pid, "HTML 2.0"):
		return "HTML 2.0"
	}

	if name == "html" {
		return "HTML (Unknown Doctype)"
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
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)

			// Determine visibility and ARIA heading metadata
			isHidden := false
			hasRoleHeading := false
			ariaLevel := 0
			for _, a := range n.Attr {
				key := strings.ToLower(a.Key)
				val := strings.ToLower(strings.TrimSpace(a.Val))
				if key == "aria-hidden" && (val == "true" || val == "1") {
					isHidden = true
				}
				if key == "hidden" {
					isHidden = true
				}
				if key == "role" && val == "heading" {
					hasRoleHeading = true
				}
				if key == "aria-level" {
					if lvl, err := strconv.Atoi(val); err == nil {
						ariaLevel = lvl
					}
				}
			}

			if !isHidden {
				// Native heading elements h1–h6
				if strings.HasPrefix(tag, "h") {
					if lvl, err := strconv.Atoi(tag[1:]); err == nil && lvl >= 1 && lvl <= 6 {
						counts[lvl]++
					}
				} else if hasRoleHeading && ariaLevel >= 1 && ariaLevel <= 6 {
					// ARIA role-based headings (count only when not a native heading)
					counts[ariaLevel]++
				}
			}
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
