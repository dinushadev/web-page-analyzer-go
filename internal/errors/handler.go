package errors

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"web-analyzer-go/internal/util"
)

// HTTPErrorHandler handles errors and sends appropriate HTTP responses
func HTTPErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Extract request ID for logging
	requestID := GetRequestID(r.Context())

	// Get or create AppError
	var appErr *AppError
	if !IsAppError(err) {
		// Convert to AppError if it's not already
		appErr = NewInternalError("An unexpected error occurred", err)
	} else {
		appErr, _ = GetAppError(err)
	}

	// Log the error with structured logging
	logFields := []any{
		slog.String("error_type", string(appErr.Type)),
		slog.String("error_message", appErr.Message),
		slog.Int("status_code", appErr.StatusCode),
		slog.String("request_id", requestID),
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
	}

	if appErr.Details != "" {
		logFields = append(logFields, slog.String("details", appErr.Details))
	}

	if appErr.Err != nil {
		logFields = append(logFields, slog.String("original_error", appErr.Err.Error()))
	}

	// Log based on severity (check if logger is initialized)
	if logger := util.EnsureLogger(); logger != nil {
		if appErr.StatusCode >= 500 {
			logger.Error("server_error", logFields...)
		} else if appErr.StatusCode >= 400 {
			logger.Warn("client_error", logFields...)
		}
	}

	// Send HTTP response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(appErr.StatusCode)

	response := appErr.ToHTTPResponse()
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the error response, log it
		if logger := util.EnsureLogger(); logger != nil {
			logger.Error("failed to encode error response",
				slog.String("encoding_error", err.Error()),
				slog.String("request_id", requestID),
			)
		}
	}
}

// Must is a helper that panics if the error is not nil
// Use this only in initialization code where errors are fatal
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// LogAndReturn logs an error and returns it
// Useful for adding logging to error returns without changing the flow
func LogAndReturn(err error, message string, fields ...any) error {
	if err == nil {
		return nil
	}

	allFields := append([]any{slog.String("error", err.Error())}, fields...)
	if logger := util.EnsureLogger(); logger != nil {
		logger.Error(message, allFields...)
	}
	return err
}

// RequestIDKey is the context key for request IDs
type RequestIDKey struct{}

// GetRequestID extracts the request ID from the context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return reqID
	}
	return ""
}
