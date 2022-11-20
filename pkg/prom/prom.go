package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const appName = "quadis"

var RoomsTotal = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: appName,
	Name:      "rooms_total",
	Help:      "Total number of currently active rooms",
})

var TotalGames = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: appName,
	Name:      "games_total",
	Help:      "Total number of games",
})
