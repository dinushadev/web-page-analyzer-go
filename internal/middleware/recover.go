package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	appErr "web-analyzer-go/internal/errors"
	"web-analyzer-go/internal/util"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())
				util.Logger.Error("panic",
					"err", rec,
					"stack", stack,
					"request_id", GetReqID(r.Context()),
				)
				// Create a standardized error for panics
				err := appErr.NewInternalError(
					fmt.Sprintf("Internal server error: %v", rec),
					fmt.Errorf("panic: %v\n%s", rec, stack),
				)
				appErr.HTTPErrorHandler(w, r, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
