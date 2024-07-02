package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	info     *prometheus.GaugeVec
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	ns := "userService"
	m := &Metrics{
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "info",
			Help:      "current running app version",
		}, []string{"version"}),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "reqiests_counter",
			Help:      "amount of served requests",
		}, []string{"version"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "request_duration_seconds",
			Help:      "duration of requests in seconds",
			Buckets:   []float64{0.02, 0.05, 0.1, 0.2, 0.5, 1, 10},
		}, []string{"version"}),
	}

	reg.MustRegister(m.info, m.requests, m.duration)

	m.info.With(prometheus.Labels{"version": "1.0.0"}).Set(1)

	return m
}

func NewMetricsMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			next.ServeHTTP(w, r)

			m.requests.With(prometheus.Labels{"version": "1.0.0"}).Inc()
			dur := time.Since(now)
			m.duration.With(prometheus.Labels{"version": "1.0.0"}).Observe(dur.Seconds())
		})
	}
}
