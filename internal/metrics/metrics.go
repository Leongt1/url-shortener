package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ShortensTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "urlshortener_shortens_total",
			Help: "Total number of times url is shortened",
		},
	)

	RedirectsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "urlshortener_redirects_total",
			Help: "Total number of redirects done",
		},
	)

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "urlshortener_http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	RedisDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "urlshortener_redis_operation_duration_seconds",
		Help:    "Duration of each redis operation in seconds",
		Buckets: []float64{0.00005, 0.0001, 0.0003, 0.0005, 0.001, 0.002, 0.005, 0.01, 0.02, 0.05},
	}, []string{"operation"})
)

func RegisterQueueDepth(fn func() float64) {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "urlshortener_click_queue_depth",
		Help: "Number of click events waiting to be processed",
	}, fn)
}
