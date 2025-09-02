package analyzer

import (
	apperrors "web-analyzer-go/internal/errors"
)

// Re-export errors from the centralized error package for backward compatibility
var (
	ErrUnreachable = apperrors.ErrUnreachable
	ErrInvalidURL  = apperrors.ErrInvalidURL
	ErrUpstream    = apperrors.ErrUpstream
	ErrParseHTML   = apperrors.ErrParseHTML
	ErrTimeout     = apperrors.ErrTimeout
)
