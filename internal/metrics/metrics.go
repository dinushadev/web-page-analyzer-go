package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	strategyDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "analyzer_strategy_duration_seconds",
			Help:    "Duration of analyzer strategies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"strategy"},
	)

	analyzeTotalDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analyzer_total_duration_seconds",
			Help:    "Total duration of analyze tasks in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func RegisterPrometheus() {
	prometheus.MustRegister(strategyDuration)
	prometheus.MustRegister(analyzeTotalDuration)
}

func Handler() http.Handler {
	return promhttp.Handler()
}

// ObserveStrategyDuration records the duration of a single strategy execution.
func ObserveStrategyDuration(strategy string, d time.Duration) {
	strategyDuration.WithLabelValues(strategy).Observe(d.Seconds())
}

// ObserveAnalyzeTotalDuration records the total duration of a full analyze task.
func ObserveAnalyzeTotalDuration(d time.Duration) {
	analyzeTotalDuration.Observe(d.Seconds())
}
