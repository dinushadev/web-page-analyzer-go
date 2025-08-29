package service

import (
	"context"
	"testing"
)

func TestAnalyzePage_SharedShim_InvalidURL(t *testing.T) {
	_, err := AnalyzePage(context.Background(), ":bad-url:")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
