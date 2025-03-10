package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace = "new_api"
)

func RegisterMetrics(registry prometheus.Registerer) {
	// channel
	registry.MustRegister(relayRequestTotalCounter)
	registry.MustRegister(relayRequestSuccessCounter)
	registry.MustRegister(relayRequestFailedCounter)
	registry.MustRegister(relayRequestRetryCounter)
	registry.MustRegister(relayRequestDurationObsever)
	// e2e
	registry.MustRegister(relayRequestE2ETotalCounter)
	registry.MustRegister(relayRequestE2ESuccessCounter)
	registry.MustRegister(relayRequestE2EFailedCounter)
	registry.MustRegister(relayRequestE2EDurationObsever)
}

var (
	relayRequestTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_total",
			Help:      "Total number of relay request total",
		}, []string{"channel", "model", "group"})
	relayRequestSuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_success",
			Help:      "Total number of relay request success",
		}, []string{"channel", "model", "group"})
	relayRequestFailedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_failed",
			Help:      "Total number of relay request failed",
		}, []string{"channel", "model", "group", "code"})
	relayRequestRetryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_retry",
			Help:      "Total number of relay request retry",
		}, []string{"channel", "model", "group"})
	relayRequestDurationObsever = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: Namespace,
			Name:      "relay_request_duration",
			Help:      "Duration of relay request",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 12),
		},
		[]string{"channel", "model", "group"},
	)
	relayRequestE2ETotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_e2e_total",
			Help:      "Total number of relay request e2e total",
		}, []string{"channel", "model", "group"})
	relayRequestE2ESuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_e2e_success",
			Help:      "Total number of relay request e2e success",
		}, []string{"channel", "model", "group"})
	relayRequestE2EFailedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_e2e_failed",
			Help:      "Total number of relay request e2e failed",
		}, []string{"channel", "model", "group", "code"})
	relayRequestE2EDurationObsever = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: Namespace,
			Name:      "relay_request_duration",
			Help:      "Duration of relay request e2e",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 12),
		},
		[]string{"channel", "model", "group"},
	)
)

func IncrementRelayRequestTotalCounter(channel, model, group string, add float64) {
	relayRequestTotalCounter.WithLabelValues(channel, model, group).Add(add)
}

func IncrementRelayRequestSuccessCounter(channel, model, group string, add float64) {
	relayRequestSuccessCounter.WithLabelValues(channel, model, group).Add(add)
}

func IncrementRelayRequestFailedCounter(channel, model, group, code string, add float64) {
	relayRequestFailedCounter.WithLabelValues(channel, model, group, code).Add(add)
}

func IncrementRelayRetryCounter(channel, model, group string, add float64) {
	relayRequestRetryCounter.WithLabelValues(channel, model, group).Add(add)
}

func ObserveRelayRequestDuration(channel, model, group string, duration float64) {
	relayRequestDurationObsever.WithLabelValues(channel, model, group).Observe(duration)
}

func IncrementRelayRequestE2ETotalCounter(channel, model, group string, add float64) {
	relayRequestE2ETotalCounter.WithLabelValues(channel, model, group).Add(add)
}

func IncrementRelayRequestE2ESuccessCounter(channel, model, group string, add float64) {
	relayRequestE2ESuccessCounter.WithLabelValues(channel, model, group).Add(add)
}

func IncrementRelayRequestE2EFailedCounter(channel, model, group, code string, add float64) {
	relayRequestE2EFailedCounter.WithLabelValues(channel, model, group, code).Add(add)
}

func ObserveRelayRequestE2EDuration(channel, model, group string, duration float64) {
	relayRequestE2EDurationObsever.WithLabelValues(channel, model, group).Observe(duration)
}
