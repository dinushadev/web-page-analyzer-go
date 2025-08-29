package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"test-project-go/internal/model"

	"golang.org/x/net/html"
)

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
		logError("http.request_build_failed", slog.String("url", targetURL), slog.String("error", reqErr.Error()))
		return nil, fmt.Errorf("build request: %w", reqErr)
	}
	resp, err := client.Do(req)
	if err != nil {
		logError("http.error", slog.String("url", targetURL), slog.String("error", ErrUnreachable.Error()))
		return nil, fmt.Errorf("get %s: %w", targetURL, ErrUnreachable)
	}
	defer resp.Body.Close()
	logInfo("http.response", slog.Int("status", resp.StatusCode), slog.String("status_text", resp.Status))
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logError("http.non_2xx", slog.Int("status", resp.StatusCode), slog.String("status_text", resp.Status))
		return nil, fmt.Errorf("upstream http %d %s: %w", resp.StatusCode, resp.Status, ErrUpstream)
	}
	logInfo("html.parse.start")
	doc, err := html.Parse(resp.Body)
	if err != nil {
		logError("html.parse.error", slog.String("error", ErrParseHTML.Error()))
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
				logError("strategy.error", slog.String("type", typeName), slog.String("error", err.Error()))
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
		logError("analyze.error", slog.String("error", "strategy error"))
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
