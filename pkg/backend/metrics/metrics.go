package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Registry is a pointer to the Prometheus Registry.
	Registry *prometheus.Registry

	// TestMetric is a dummy metric example.
	TestMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

// RegisterAll is a wrapper function for the universal Prometheus metric registration.
func RegisterAll() {
	Registry = prometheus.NewRegistry()

	Registry.MustRegister(TestMetric)
}
