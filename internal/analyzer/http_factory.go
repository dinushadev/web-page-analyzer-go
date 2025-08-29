package analyzer

import (
	"net/http"
	"time"
)

type HTTPClientFactory interface {
	NewClient() *http.Client
}

type DefaultHTTPClientFactory struct{}

func (f *DefaultHTTPClientFactory) NewClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

type LinkChecker interface {
	IsAccessible(link string) bool
}

type DefaultLinkChecker struct {
	Client *http.Client
}

func (c *DefaultLinkChecker) IsAccessible(link string) bool {
	resp, err := c.Client.Head(link)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
