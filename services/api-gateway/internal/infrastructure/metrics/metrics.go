package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry     *prometheus.Registry
	httpRequests *prometheus.CounterVec
	httpLatency  *prometheus.HistogramVec
}

func New(serviceName string) *Metrics {
	registry := prometheus.NewRegistry()

	httpRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "sporttech_http_requests_total",
			Help:        "Total number of HTTP requests handled by the service.",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"method", "path", "status"},
	)
	httpLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "sporttech_http_request_duration_seconds",
			Help:        "HTTP request latency in seconds.",
			ConstLabels: prometheus.Labels{"service": serviceName},
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	registry.MustRegister(
		httpRequests,
		httpLatency,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &Metrics{
		registry:     registry,
		httpRequests: httpRequests,
		httpLatency:  httpLatency,
	}
}

func (metrics *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{})
}

func (metrics *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		startedAt := time.Now()
		next.ServeHTTP(recorder, request)

		statusCode := strconv.Itoa(recorder.statusCode)
		metrics.httpRequests.WithLabelValues(request.Method, request.URL.Path, statusCode).Inc()
		metrics.httpLatency.WithLabelValues(request.Method, request.URL.Path, statusCode).Observe(time.Since(startedAt).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (recorder *statusRecorder) WriteHeader(statusCode int) {
	recorder.statusCode = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}
