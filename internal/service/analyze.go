package service

import (
	"context"
	"test-project-go/internal/analyzer"
	"test-project-go/internal/model"
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
