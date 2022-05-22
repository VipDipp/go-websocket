package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RegisteredUsersCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "registered_users_total",
		Help: "The total number of registered users",
	})

	GivenCakesCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "given_cakes_total",
		Help: "The total number of given cakes",
	})

	UserInfoCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "user_info_total",
		Help: "The total number of user info giving",
	})

	buckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

	responseTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "namespace",
		Name:      "http_server_request_duration_seconds",
		Help:      "Histogram of response time for handler in seconds",
		Buckets:   buckets,
	}, []string{"route", "method", "status_code"})
)

//statusRecorder to record the status code from the ResponseWriter
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func measureResponseDuration(next ProtectedHandler) ProtectedHandler {
	return func(w http.ResponseWriter, r *http.Request, u User, users UserRepository) {
		start := time.Now()
		rec := statusRecorder{w, 200}

		next(w, r, u, users)

		duration := time.Since(start)
		statusCode := strconv.Itoa(rec.statusCode)
		responseTimeHistogram.WithLabelValues(r.URL.Path, r.Method, statusCode).Observe(duration.Seconds())
	}
}

// getRoutePattern returns the route pattern from the chi context there are 3 conditions
// a) static routes "/example" => "/example"
// b) dynamic routes "/example/:id" => "/example/{id}"
// c) if nothing matches the output is undefined

func PrometheusRun() {
	prometheus.MustRegister(responseTimeHistogram)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
