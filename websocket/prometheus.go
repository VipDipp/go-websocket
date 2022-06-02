package websocket

import (
	"net/http"

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

	ResponseTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "namespace",
		Name:      "http_server_request_duration_seconds",
		Help:      "Histogram of response time for handler in seconds",
		Buckets:   buckets,
	}, []string{"route", "method", "status_code"})
)

//statusRecorder to record the status code from the ResponseWriter
type StatusRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func (rec *StatusRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// getRoutePattern returns the route pattern from the chi context there are 3 conditions
// a) static routes "/example" => "/example"
// b) dynamic routes "/example/:id" => "/example/{id}"
// c) if nothing matches the output is undefined

func PrometheusRun() {
	prometheus.MustRegister(ResponseTimeHistogram)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
