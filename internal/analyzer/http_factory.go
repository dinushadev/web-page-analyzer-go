package analyzer

import (
	"context"
	"net"
	"net/http"
	"time"
)

const userAgent = "test-project-go/1.0"

type HTTPClientFactory interface {
	NewClient() *http.Client
}

type DefaultHTTPClientFactory struct{}

func (f *DefaultHTTPClientFactory) NewClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 120 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

type LinkChecker interface {
	IsAccessible(link string) bool
}

type DefaultLinkChecker struct {
	Client *http.Client
}

func (c *DefaultLinkChecker) IsAccessible(link string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try HEAD first with User-Agent
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, link, nil)
	if err == nil {
		req.Header.Set("User-Agent", userAgent)
		if resp, doErr := c.Client.Do(req); doErr == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				return true
			}
			// If HEAD not supported, fall through to GET
			if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusNotImplemented {
				return false
			}
		}
	}

	// Fallback: GET with Range to minimize payload
	reqGet, err2 := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err2 != nil {
		return false
	}
	reqGet.Header.Set("User-Agent", userAgent)
	reqGet.Header.Set("Range", "bytes=0-0")
	respGet, errDo := c.Client.Do(reqGet)
	if errDo != nil {
		return false
	}
	defer respGet.Body.Close()
	return respGet.StatusCode >= 200 && respGet.StatusCode < 400
}
