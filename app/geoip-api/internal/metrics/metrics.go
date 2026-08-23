package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoip_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "geoip_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "geoip_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "geoip_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	ExternalAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoip_external_api_calls_total",
			Help: "Total number of external GeoIP API calls",
		},
		[]string{"status"},
	)

	ExternalAPIDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "geoip_external_api_duration_seconds",
			Help:    "External GeoIP API call duration in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
	)
)

func Register() {}
