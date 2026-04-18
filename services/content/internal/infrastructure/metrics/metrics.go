package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Metrics struct {
	registry     *prometheus.Registry
	httpRequests *prometheus.CounterVec
	httpLatency  *prometheus.HistogramVec
	grpcRequests *prometheus.CounterVec
	grpcLatency  *prometheus.HistogramVec
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
	grpcRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "sporttech_grpc_requests_total",
			Help:        "Total number of gRPC requests handled by the service.",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"method", "code"},
	)
	grpcLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "sporttech_grpc_request_duration_seconds",
			Help:        "gRPC request latency in seconds.",
			ConstLabels: prometheus.Labels{"service": serviceName},
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)

	registry.MustRegister(
		httpRequests,
		httpLatency,
		grpcRequests,
		grpcLatency,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &Metrics{
		registry:     registry,
		httpRequests: httpRequests,
		httpLatency:  httpLatency,
		grpcRequests: grpcRequests,
		grpcLatency:  grpcLatency,
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
		path := recorder.routePattern
		if path == "" {
			path = request.Pattern
		}
		if path == "" {
			path = request.URL.Path
		}

		metrics.httpRequests.WithLabelValues(request.Method, path, statusCode).Inc()
		metrics.httpLatency.WithLabelValues(request.Method, path, statusCode).Observe(time.Since(startedAt).Seconds())
	})
}

func (metrics *Metrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		startedAt := time.Now()
		response, err := handler(ctx, request)

		code := status.Code(err).String()
		metrics.grpcRequests.WithLabelValues(info.FullMethod, code).Inc()
		metrics.grpcLatency.WithLabelValues(info.FullMethod, code).Observe(time.Since(startedAt).Seconds())

		return response, err
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	routePattern string
}

func (recorder *statusRecorder) WriteHeader(statusCode int) {
	recorder.statusCode = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}

func (recorder *statusRecorder) SetRoutePattern(routePattern string) {
	recorder.routePattern = routePattern
}
