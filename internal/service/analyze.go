package service

import (
	"context"
	"web-analyzer-go/internal/analyzer"
	"web-analyzer-go/internal/model"
)

var (
	ErrUnreachable = analyzer.ErrUnreachable
	ErrInvalidURL  = analyzer.ErrInvalidURL
	ErrUpstream    = analyzer.ErrUpstream
	ErrParseHTML   = analyzer.ErrParseHTML
	ErrTimeout     = analyzer.ErrTimeout
)

func AnalyzePage(ctx context.Context, url string) (*model.AnalyzeResult, error) {
	return analyzer.AnalyzePage(ctx, url)
}
