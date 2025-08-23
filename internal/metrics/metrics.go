package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func RegisterPrometheus() {
	// Register custom metrics here if needed
}

func Handler() http.Handler {
	return promhttp.Handler()
}
