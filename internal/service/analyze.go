package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"test-project-go/internal/model"
	"test-project-go/internal/util"

	"golang.org/x/net/html"
)

var ErrUnreachable = errors.New("url is unreachable")
var ErrInvalidURL = errors.New("invalid url")
var ErrUpstream = errors.New("upstream http error")
var ErrParseHTML = errors.New("failed to parse html")

func logInfo(message string, args ...any) {
	if util.Logger != nil {
		util.Logger.Info(message, args...)
	}
}

func mergeAnalyzeResult(main, partial *model.AnalyzeResult) {
	if partial.HTMLVersion != "" {
		main.HTMLVersion = partial.HTMLVersion
	}
	if partial.Title != "" {
		main.Title = partial.Title
	}
	if len(partial.Headings) > 0 {
		main.Headings = partial.Headings
	}
	if partial.Links.Internal != 0 || partial.Links.External != 0 || partial.Links.Inaccessible != 0 {
		main.Links = partial.Links
	}
	if partial.LoginForm {
		main.LoginForm = true
	}
}

func AnalyzePage(ctx context.Context, targetURL string) (*model.AnalyzeResult, error) {
	logInfo("analyze.start", slog.String("url", targetURL))
	parsed, err := url.ParseRequestURI(targetURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		logInfo("analyze.invalid_url", slog.String("url", targetURL))
		return nil, fmt.Errorf("parse url %q: %w", targetURL, ErrInvalidURL)
	}
	logInfo("analyze.url_parsed", slog.String("host", parsed.Host))

	clientFactory := &DefaultHTTPClientFactory{}
	client := clientFactory.NewClient()
	logInfo("http.fetch", slog.String("url", targetURL))
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("build request: %w", reqErr)
	}
	resp, err := client.Do(req)
	if err != nil {
		logInfo("http.error", slog.String("url", targetURL), slog.String("error", ErrUnreachable.Error()))
		return nil, fmt.Errorf("get %s: %w", targetURL, ErrUnreachable)
	}
	defer resp.Body.Close()
	logInfo("http.response", slog.Int("status", resp.StatusCode), slog.String("status_text", resp.Status))
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logInfo("http.non_2xx", slog.Int("status", resp.StatusCode), slog.String("status_text", resp.Status))
		return nil, fmt.Errorf("upstream http %d %s: %w", resp.StatusCode, resp.Status, ErrUpstream)
	}
	logInfo("html.parse.start")
	doc, err := html.Parse(resp.Body)
	if err != nil {
		logInfo("html.parse.error", slog.String("error", ErrParseHTML.Error()))
		return nil, fmt.Errorf("parse html: %w", ErrParseHTML)
	}
	logInfo("html.parse.ok")

	strategies := []AnalyzerStrategy{
		&HTMLVersionStrategy{},
		&TitleStrategy{},
		&HeadingsStrategy{},
		&LinksStrategy{LinkChecker: &DefaultLinkChecker{Client: client}},
		&LoginFormStrategy{},
	}
	result := &model.AnalyzeResult{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(strategies))

	logInfo("strategies.start", slog.Int("count", len(strategies)))
	for _, s := range strategies {
		wg.Add(1)
		go func(strategy AnalyzerStrategy) {
			defer wg.Done()
			partial := &model.AnalyzeResult{}
			typeName := reflect.TypeOf(strategy).String()
			logInfo("strategy.start", slog.String("type", typeName))
			if err := strategy.Analyze(doc, parsed, partial); err != nil {
				errChan <- err
				return
			}
			mu.Lock()
			mergeAnalyzeResult(result, partial)
			mu.Unlock()
			logInfo("strategy.done", slog.String("type", typeName),
				slog.String("html_version", partial.HTMLVersion),
				slog.String("title", partial.Title),
				slog.Bool("login_form", partial.LoginForm),
				slog.Int("links_internal", partial.Links.Internal),
				slog.Int("links_external", partial.Links.External),
				slog.Int("links_inaccessible", partial.Links.Inaccessible),
			)
		}(s)
	}
	wg.Wait()
	logInfo("strategies.done")
	close(errChan)
	if len(errChan) > 0 {
		logInfo("analyze.error", slog.String("error", "strategy error"))
		return nil, fmt.Errorf("strategy error: %w", <-errChan)
	}
	logInfo("analyze.done",
		slog.String("html_version", result.HTMLVersion),
		slog.String("title", result.Title),
		slog.Bool("login_form", result.LoginForm),
		slog.Int("links_internal", result.Links.Internal),
		slog.Int("links_external", result.Links.External),
		slog.Int("links_inaccessible", result.Links.Inaccessible),
	)
	return result, nil
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
		result = append(result, model.HeadingCount{Level: i, Count: counts[i]})
	}
	return result
}

// countLinks recursively finds and counts links in the HTML node tree.
func countLinks(n *html.Node, base *url.URL, checker LinkChecker) (internal, external, inaccessible int) {
	// A nested recursive function to process nodes.
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			link := ""
			// Check for href and src attributes in various tags.
			for _, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					link = attr.Val
					break
				}
			}

			if link != "" {
				// Step 1: Parse and validate the link.
				u, err := url.Parse(link)
				if err != nil || u.Scheme == "javascript" || u.Scheme == "mailto" {
					return
				}

				// Step 2: Resolve to an absolute URL.
				abs := u
				if !u.IsAbs() {
					abs = base.ResolveReference(u)
				}

				// Step 3: Check for accessibility first.
				if !checker.IsAccessible(abs.String()) {
					inaccessible++
					return // Stop processing this link if it's not accessible.
				}

				// Step 4: Categorize the link.
				if abs.Host == base.Host {
					internal++
				} else {
					external++
				}
			}
		}

		// Recurse through the children of the current node.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	// Start the recursive process from the root node.
	f(n)
	return
}

// hasLoginForm recursively searches for a login form in the HTML node tree.
func hasLoginForm(n *html.Node) bool {
	// First, check if the current node is a form.
	if n.Type == html.ElementNode && n.Data == "form" {
		// If it's a form, we'll check its children for a password input.
		if containsPasswordInput(n) {
			logInfo("login_form.detected")
			return true
		}
	}

	// Now, recursively check the children of the current node.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasLoginForm(c) {
			return true
		}
	}
	return false
}

// containsPasswordInput recursively checks if a node or its children
// contain an input field with type="password".
func containsPasswordInput(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "input" {
		for _, attr := range n.Attr {
			// A case-insensitive check is more robust.
			if strings.EqualFold(attr.Key, "type") && strings.EqualFold(attr.Val, "password") {
				return true
			}
		}
	}
	// Continue the search through the node's children.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if containsPasswordInput(c) {
			return true
		}
	}
	return false
}
