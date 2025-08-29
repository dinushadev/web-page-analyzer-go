package service

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestDetectHTMLVersion(t *testing.T) {
	html5 := `<!DOCTYPE html><html><head><title>Test</title></head><body></body></html>`
	doc, _ := html.Parse(strings.NewReader(html5))
	v := detectHTMLVersion(doc)
	if v != "HTML5" {
		t.Errorf("expected HTML5, got %s", v)
	}
}

func TestExtractTitle(t *testing.T) {
	h := `<!DOCTYPE html><html><head><title>My Page</title></head><body></body></html>`
	doc, _ := html.Parse(strings.NewReader(h))
	title := extractTitle(doc)
	if title != "My Page" {
		t.Errorf("expected 'My Page', got '%s'", title)
	}
}

func TestCountHeadings(t *testing.T) {
	h := `<!DOCTYPE html><html><body><h1>H1</h1><h2>H2</h2><h2>H2b</h2></body></html>`
	doc, _ := html.Parse(strings.NewReader(h))
	headings := countHeadings(doc)
	if len(headings) != 6 {
		t.Fatalf("expected 6 heading levels, got %d: %+v", len(headings), headings)
	}
	expected := []struct{ level, count int }{{1, 1}, {2, 2}, {3, 0}, {4, 0}, {5, 0}, {6, 0}}
	for i, e := range expected {
		if headings[i].Level != e.level || headings[i].Count != e.count {
			t.Errorf("unexpected h%d count: got %d", e.level, headings[i].Count)
		}
	}
}

func TestHasLoginForm(t *testing.T) {
	h := `<!DOCTYPE html><html><body><form><input type="password"/></form></body></html>`
	doc, _ := html.Parse(strings.NewReader(h))
	if !hasLoginForm(doc) {
		t.Error("expected login form to be detected")
	}
}

func TestAnalyzePage_InvalidURL(t *testing.T) {
	_, err := AnalyzePage(context.Background(), ":bad-url:")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestAnalyzePage_Unreachable(t *testing.T) {
	// Use a non-routable IP to simulate unreachable
	_, err := AnalyzePage(context.Background(), "http://10.255.255.1")
	if err == nil {
		t.Error("expected unreachable error")
	}
}

// Use a mock LinkChecker that always returns true (all links accessible)
type mockChecker struct{}

func (m *mockChecker) IsAccessible(link string) bool { return true }

func TestCountLinks(t *testing.T) {
	h := `<!DOCTYPE html><html><body>
	<a href="/internal">Internal</a>
	<a href="http://external.com">External</a>
	<a href="mailto:test@example.com">Mail</a>
	</body></html>`
	base, _ := http.NewRequest("GET", "http://localhost", nil)
	doc, _ := html.Parse(strings.NewReader(h))
	internal, external, _ := countLinks(doc, base.URL, &mockChecker{})
	if internal != 1 || external != 1 {
		t.Errorf("expected 1 internal and 1 external, got %d internal, %d external", internal, external)
	}
	// Inaccessible count may vary depending on network, so we don't assert it here
}
