package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"test-project-go/internal/factory"
	"test-project-go/internal/model"
	"testing"

	"golang.org/x/net/html"
)

func Test_parseTargetURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
		ok   bool
	}{
		{"http", "http://example.com", true},
		{"https", "https://example.com", true},
		{"invalid scheme", "ftp://example.com", false},
		{"garbage", "://", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := parseTargetURL(tc.url)
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected error, got ok: %v", u)
			}
		})
	}
}

func Test_buildGetRequest_setsUserAgent(t *testing.T) {
	req, err := buildGetRequest(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := req.Header.Get("User-Agent"); got == "" {
		t.Fatalf("expected User-Agent to be set")
	}
}

func Test_fetchAndParseHTML_non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	}))
	defer srv.Close()

	client := (&factory.DefaultHTTPClientFactory{}).NewClient()
	req, _ := buildGetRequest(context.Background(), srv.URL)
	_, err := fetchAndParseHTML(client, req)
	if err == nil {
		t.Fatalf("expected error for non-2xx status")
	}
}

type strategyOK struct{}

func (s *strategyOK) Analyze(_ *html.Node, _ *url.URL, r *model.AnalyzeResult) error {
	r.Title = "ok"
	return nil
}

type strategyErr struct{}

func (s *strategyErr) Analyze(_ *html.Node, _ *url.URL, _ *model.AnalyzeResult) error {
	return fmt.Errorf("boom")
}

type strategyLogin struct{}

func (s *strategyLogin) Analyze(_ *html.Node, _ *url.URL, r *model.AnalyzeResult) error {
	r.LoginForm = true
	return nil
}

func Test_runStrategiesParallel_merge(t *testing.T) {
	doc, _ := html.Parse(strings.NewReader("<!DOCTYPE html><html><head><title>X</title></head><body></body></html>"))
	base, _ := url.Parse("http://example.com")
	res, err := runStrategiesParallel(context.Background(), doc, base, []AnalyzerStrategy{&strategyOK{}, &strategyLogin{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "ok" || !res.LoginForm {
		t.Fatalf("unexpected merge result: %+v", res)
	}
}

func Test_runStrategiesParallel_error(t *testing.T) {
	doc, _ := html.Parse(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	base, _ := url.Parse("http://example.com")
	_, err := runStrategiesParallel(context.Background(), doc, base, []AnalyzerStrategy{&strategyOK{}, &strategyErr{}})
	if err == nil {
		t.Fatalf("expected error from failing strategy")
	}
}

func Test_AnalyzePage_integration_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<!DOCTYPE html><html><head><title>T</title></head><body><a href='" + srvURL(w, r) + "'>x</a></body></html>"))
	}))
	defer srv.Close()

	res, err := AnalyzePage(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "T" {
		t.Fatalf("unexpected title: %+v", res)
	}
}

// helper to get absolute server URL
func srvURL(w http.ResponseWriter, r *http.Request) string {
	return "http://" + r.Host
}

func Test_DefaultLinkChecker_HEAD_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Fatalf("unexpected method: %s", r.Method)
	}))
	defer srv.Close()

	checker := &factory.DefaultLinkChecker{Client: (&factory.DefaultHTTPClientFactory{}).NewClient()}
	if ok := checker.IsAccessible(srv.URL); !ok {
		t.Fatalf("expected accessible for HEAD 200")
	}
}

func Test_DefaultLinkChecker_HEAD_405_GET_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusMethodNotAllowed)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer srv.Close()

	checker := &factory.DefaultLinkChecker{Client: (&factory.DefaultHTTPClientFactory{}).NewClient()}
	if ok := checker.IsAccessible(srv.URL); !ok {
		t.Fatalf("expected accessible for HEAD 405 then GET 200")
	}
}

func Test_DefaultLinkChecker_HEAD_405_GET_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusMethodNotAllowed)
		case http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer srv.Close()

	checker := &factory.DefaultLinkChecker{Client: (&factory.DefaultHTTPClientFactory{}).NewClient()}
	if ok := checker.IsAccessible(srv.URL); ok {
		t.Fatalf("expected inaccessible for HEAD 405 then GET 404")
	}
}

func Test_DefaultLinkChecker_HEAD_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		t.Fatalf("unexpected method: %s", r.Method)
	}))
	defer srv.Close()

	checker := &factory.DefaultLinkChecker{Client: (&factory.DefaultHTTPClientFactory{}).NewClient()}
	if ok := checker.IsAccessible(srv.URL); ok {
		t.Fatalf("expected inaccessible for HEAD 404")
	}
}
