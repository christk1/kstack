package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var reqs = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP requests",
	},
	[]string{"path"},
)

func main() {
	prometheus.MustRegister(reqs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqs.WithLabelValues(r.URL.Path).Inc()
		fmt.Fprintln(w, "Hello from example-app")
	})
	addr := ":" + port
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
