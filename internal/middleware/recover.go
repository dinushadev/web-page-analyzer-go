package middleware

import (
	"net/http"
	"runtime/debug"
	"test-project-go/internal/util"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				util.Logger.Error("panic",
					"err", rec,
					"stack", string(debug.Stack()),
					"request_id", GetReqID(r.Context()),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
