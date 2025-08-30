package analyzer

import (
	"net/http"
	"strings"
	"testing"
	"web-analyzer-go/internal/util"

	"golang.org/x/net/html"
)

func TestDetectHTMLVersion(t *testing.T) {
	html5 := `<!DOCTYPE html><html><head><title>Test</title></head><body></body></html>`
	doc, _ := html.Parse(strings.NewReader(html5))
	v := util.DetectHTMLVersion(doc)
	if v != "HTML5" {
		t.Errorf("expected HTML5, got %s", v)
	}
}

func TestDetectHTMLVersion_HTML4(t *testing.T) {
	html4 := `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
	<html><head><title>Test</title></head><body></body></html>`
	doc, _ := html.Parse(strings.NewReader(html4))
	v := util.DetectHTMLVersion(doc)
	if v != "HTML 4.01 Transitional" {
		t.Errorf("expected HTML 4.01 Transitional, got %s", v)
	}
}

func TestDetectHTMLVersion_HTML2(t *testing.T) {
	html2 := `<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN">
	<html><head><title>Test</title></head><body></body></html>`
	doc, _ := html.Parse(strings.NewReader(html2))
	v := util.DetectHTMLVersion(doc)
	if v != "HTML 2.0" {
		t.Errorf("expected HTML 2.0, got %s", v)
	}
}

func TestExtractTitle(t *testing.T) {
	h := `<!DOCTYPE html><html><head><title>My Page</title></head><body></body></html>`
	doc, _ := html.Parse(strings.NewReader(h))
	title := util.ExtractTitle(doc)
	if title != "My Page" {
		t.Errorf("expected 'My Page', got '%s'", title)
	}
}

func TestCountHeadings(t *testing.T) {
	h := `<!DOCTYPE html><html><body><h1>H1</h1><h2>H2</h2><h2>H2b</h2></body></html>`
	doc, _ := html.Parse(strings.NewReader(h))
	headings := util.CountHeadings(doc)
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
	if !util.HasLoginForm(doc) {
		t.Error("expected login form to be detected")
	}
}

// Use a mock LinkChecker that always returns true (all links accessible)
type mockChecker struct{}

func (m *mockChecker) IsAccessible(link string) bool { return true }

func TestCountLinks(t *testing.T) {
	h := `<!DOCTYPE html><html><body>
	<a href="/internal">Internal</a>
	<a href="http://external.com">External</a>
	<a href="mailto:test@simplewebapp.com">Mail</a>
	</body></html>`
	base, _ := http.NewRequest("GET", "http://localhost", nil)
	doc, _ := html.Parse(strings.NewReader(h))
	internal, external, _ := util.CountLinks(doc, base.URL, (&mockChecker{}).IsAccessible)
	if internal != 1 || external != 1 {
		t.Errorf("expected 1 internal and 1 external, got %d internal, %d external", internal, external)
	}
}
