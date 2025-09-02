package analyzer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
	appErr "web-analyzer-go/internal/errors"
	"web-analyzer-go/internal/factory"
	"web-analyzer-go/internal/metrics"
	"web-analyzer-go/internal/model"

	"io"

	"golang.org/x/net/html"
	"golang.org/x/sync/errgroup"
)

// AnalyzePage fetches the target URL, parses its HTML, and runs a suite of
// analysis strategies to produce a consolidated *model.AnalyzeResult. It is
// safe for concurrent use and respects the provided context for request-level
// timeouts and cancellation.
func AnalyzePage(ctx context.Context, targetURL string) (*model.AnalyzeResult, error) {
	logInfo("analyze.start", slog.String("url", targetURL))
	start := time.Now()

	parsed, err := parseTargetURL(targetURL)
	if err != nil {
		return nil, err
	}

	client := (&factory.DefaultHTTPClientFactory{}).NewClient()

	req, err := buildGetRequest(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	doc, err := fetchAndParseHTML(client, req)
	if err != nil {
		return nil, err
	}

	result, err := runStrategiesParallel(ctx, doc, parsed, strategiesFor(client))
	if err != nil {
		return nil, err
	}

	logInfo("analyze.done",
		slog.String("html_version", result.HTMLVersion),
		slog.String("title", result.Title),
		slog.Bool("login_form", result.LoginForm),
		slog.Int("links_internal", result.Links.Internal),
		slog.Int("links_external", result.Links.External),
		slog.Int("links_inaccessible", result.Links.Inaccessible),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
	)
	metrics.ObserveAnalyzeTotalDuration(time.Since(start))
	return result, nil
}

// parseTargetURL validates and parses the provided URL.
func parseTargetURL(targetURL string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(targetURL)
	if err != nil {
		logInfo("analyze.invalid_url", slog.String("url", targetURL))
		return nil, appErr.NewValidationError("invalid URL format", err.Error())
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		logInfo("analyze.invalid_url", slog.String("url", targetURL))
		return nil, appErr.NewValidationError("URL must use http or https scheme", fmt.Sprintf("got scheme: %s", parsed.Scheme))
	}
	logInfo("analyze.url_parsed", slog.String("host", parsed.Host))
	return parsed, nil
}

// buildGetRequest constructs a GET request with the appropriate headers.
func buildGetRequest(ctx context.Context, targetURL string) (*http.Request, error) {
	logInfo("http.fetch", slog.String("url", targetURL))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		logError("http.request_build_failed", slog.String("url", targetURL), slog.String("error", err.Error()))
		return nil, appErr.NewInternalError("failed to create HTTP request", err)
	}
	req.Header.Set("User-Agent", factory.UserAgent)
	return req, nil
}

// fetchAndParseHTML executes the request and parses the response body as HTML.
func fetchAndParseHTML(client *http.Client, req *http.Request) (*html.Node, error) {
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			logError("http.timeout", slog.String("url", req.URL.String()))
			return nil, appErr.NewTimeoutError(fmt.Sprintf("request to %s", req.URL.String()), err)
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			logError("http.timeout", slog.String("url", req.URL.String()))
			return nil, appErr.NewTimeoutError(fmt.Sprintf("request to %s", req.URL.String()), err)
		}
		logError("http.error", slog.String("url", req.URL.String()), slog.String("error", ErrUnreachable.Error()))
		return nil, appErr.WrapError(err, appErr.ErrorTypeUnavailable, fmt.Sprintf("URL %s is unreachable", req.URL.String()))
	}
	defer resp.Body.Close()

	logInfo("http.response", slog.Int("status", resp.StatusCode), slog.String("status_text", resp.Status))
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		logError("http.non_2xx", slog.Int("status", resp.StatusCode), slog.String("status_text", resp.Status))
		return nil, appErr.NewUpstreamError(req.URL.Host, resp.StatusCode, fmt.Errorf("received status: %s", resp.Status))
	}

	logInfo("html.parse.start")
	limitedBody := io.LimitReader(resp.Body, 2<<20) // 2MB cap
	doc, parseErr := html.Parse(limitedBody)
	if parseErr != nil {
		logError("html.parse.error", slog.String("error", ErrParseHTML.Error()))
		return nil, appErr.NewParseError("HTML", parseErr)
	}
	logInfo("html.parse.ok")
	return doc, nil
}

// strategiesFor builds the list of analysis strategies with dependencies injected.
func strategiesFor(client *http.Client) []AnalyzerStrategy {
	return []AnalyzerStrategy{
		&HTMLVersionStrategy{},
		&TitleStrategy{},
		&HeadingsStrategy{},
		&LinksStrategy{LinkChecker: &factory.DefaultLinkChecker{Client: client}},
		&LoginFormStrategy{},
	}
}

// runStrategiesParallel executes all strategies concurrently, merging results.
// The provided context is used to cancel in-flight strategy work if any strategy fails.
func runStrategiesParallel(ctx context.Context, doc *html.Node, base *url.URL, strategies []AnalyzerStrategy) (*model.AnalyzeResult, error) {
	result := &model.AnalyzeResult{}
	var mu sync.Mutex
	group, ctx := errgroup.WithContext(ctx)

	var slowestName string
	var slowestDur time.Duration

	logInfo("strategies.start", slog.Int("count", len(strategies)))
	for _, s := range strategies {
		strategy := s
		group.Go(func() error {
			partial := &model.AnalyzeResult{}
			strategyType := fmt.Sprintf("%T", strategy)
			st := time.Now()
			logInfo("strategy.start", slog.String("type", strategyType))
			// Strategy interface does not take context; honor cancellation best-effort by early return
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if err := strategy.Analyze(doc, base, partial); err != nil {
				logError("strategy.error", slog.String("type", strategyType), slog.String("error", err.Error()))
				return err
			}
			d := time.Since(st)
			metrics.ObserveStrategyDuration(strategyType, d)
			mu.Lock()
			mergeAnalyzeResult(result, partial)
			if d > slowestDur {
				slowestDur = d
				slowestName = strategyType
			}
			mu.Unlock()
			logInfo("strategy.done", slog.String("type", strategyType),
				slog.String("html_version", partial.HTMLVersion),
				slog.String("title", partial.Title),
				slog.Bool("login_form", partial.LoginForm),
				slog.Int("links_internal", partial.Links.Internal),
				slog.Int("links_external", partial.Links.External),
				slog.Int("links_inaccessible", partial.Links.Inaccessible),
				slog.Int64("duration_ms", d.Milliseconds()),
			)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		logError("analyze.error", slog.String("error", "strategy error"))
		return nil, appErr.WrapError(err, appErr.ErrorTypeInternal, "analysis strategy failed")
	}
	logInfo("strategies.done", slog.String("slowest_strategy", slowestName), slog.Int64("slowest_ms", slowestDur.Milliseconds()))
	return result, nil
}
