package ipset

import (
	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Variables declared for monitoring.
var (
	AddIPCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "ipset",
		Name:      "add_ip_count_total",
		Help:      "Counter of add IP to ipset.",
	}, []string{"to"})
)
