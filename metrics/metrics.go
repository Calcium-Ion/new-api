package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace = "new_api"
)

func RegisterMetrics(registry prometheus.Registerer) {
	registry.MustRegister(relayRequestTotalCounter)
	registry.MustRegister(relayRequestSuccessCounter)
	registry.MustRegister(relayRequestFailedCounter)
	registry.MustRegister(relayRequestRetryCounter)
	registry.MustRegister(relayRequestDurationObsever)
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
		}, []string{"channel", "model", "group", "code", "msg"})
	relayRequestRetryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: Namespace,
			Name:      "relay_request_retry",
			Help:      "Total number of relay request retry",
		}, []string{"channel", "group"})
	relayRequestDurationObsever = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: Namespace,
			Name:      "relay_request_duration",
			Help:      "Duration of relay request",
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

func IncrementRelayRequestFailedCounter(channel, model, group, code, msg string, add float64) {
	relayRequestFailedCounter.WithLabelValues(channel, model, group, code, msg).Add(add)
}

func IncrementRelayRetryCounter(channel, group string, add float64) {
	relayRequestRetryCounter.WithLabelValues(channel, group).Add(add)
}

func ObserveRelayRequestDuration(channel, model, group string, duration float64) {
	relayRequestDurationObsever.WithLabelValues(channel, model, group).Observe(duration)
}
