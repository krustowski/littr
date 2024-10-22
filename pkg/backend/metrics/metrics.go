package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"go.vxn.dev/swis/v5/pkg/core"
)

var (
	// Registry is a pointer to the Prometheus Registry.
	Registry *prometheus.Registry

	// Metrics struct
	Metrics *metrics
)

type metrics struct {
	PollCountMetric         prometheus.Gauge
	PostCountMetric         prometheus.Gauge
	RequestCountMetric      prometheus.Gauge
	SubscriptionCountMetric prometheus.Gauge
	TokenCountMetric        prometheus.Gauge
	UserCountMetric         prometheus.Gauge
}

// RegisterAll is a wrapper function for the universal Prometheus metric registration.
func RegisterAll() {
	Registry = prometheus.NewRegistry()

	Metrics = &metrics{
		// PollCountMetric track the number of polls actually loaded in memory.
		PollCountMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "littr_poll_count_total",
			Help: "The total number of polls loaded in memory.",
		}),

		// PostCountMetric track the number of posts actually loaded in memory.
		PostCountMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "littr_post_count_total",
			Help: "The total number of posts loaded in memory.",
		}),

		// RequestCountMetric track the number of generic requests actually loaded in memory.
		RequestCountMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "littr_request_count_total",
			Help: "The total number of generic requests loaded in memory.",
		}),

		// SubscriptionCountMetric track the number of subscription actually loaded in memory.
		SubscriptionCountMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "littr_subscription_count_total",
			Help: "The total number of subscriptions loaded in memory.",
		}),

		// TokenCountMetric track the number of tokens actually loaded in memory.
		TokenCountMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "littr_token_count_total",
			Help: "The total number of refresh token references loaded in memory.",
		}),

		// UserCountMetric track the number of users actually loaded in memory.
		UserCountMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "littr_user_count_total",
			Help: "The total number of users loaded in memory.",
		}),
	}

	//m := eng.NewMetrics(reg)
	//m.One.Set(888)
	//m.Two.With(prometheus.Labels{"device": "custom name"}).Inc()

	Registry.MustRegister(Metrics.PollCountMetric, Metrics.PostCountMetric, Metrics.RequestCountMetric, Metrics.SubscriptionCountMetric, Metrics.TokenCountMetric, Metrics.UserCountMetric)
}

func UpdateCountMetric(cache *core.Cache, count int, absolute bool) {
	switch cache.Name {
	case "FlowCache":
		if absolute {
			Metrics.PostCountMetric.Set(float64(count))
			return
		}

		Metrics.PostCountMetric.Add(float64(count))
	case "PollCache":
		if absolute {
			Metrics.PollCountMetric.Set(float64(count))
			return
		}

		Metrics.PollCountMetric.Add(float64(count))
	case "RequestCache":
		if absolute {
			Metrics.RequestCountMetric.Set(float64(count))
			return
		}

		Metrics.RequestCountMetric.Add(float64(count))
	case "SubscriptionCache":
		if absolute {
			Metrics.SubscriptionCountMetric.Set(float64(count))
			return
		}

		Metrics.SubscriptionCountMetric.Add(float64(count))
	case "TokenCache":
		if absolute {
			Metrics.TokenCountMetric.Set(float64(count))
			return
		}

		Metrics.TokenCountMetric.Add(float64(count))
	case "UserCache":
		if absolute {
			Metrics.UserCountMetric.Set(float64(count))
			return
		}

		Metrics.UserCountMetric.Add(float64(count))
	default:
	}
}
