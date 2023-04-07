package metrics

import (
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
)

var pieceTokenNames = map[piece.Token]string{
	piece.TokenI:       "I",
	piece.TokenO:       "O",
	piece.TokenL:       "L",
	piece.TokenJ:       "J",
	piece.TokenS:       "S",
	piece.TokenT:       "T",
	piece.TokenZ:       "Z",
	piece.TokenBedrock: "Bedrock",
}

const appName = "quadis"

var RoomsTotal = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: appName,
	Name:      "rooms_total",
	Help:      "Total number of currently active rooms",
})

var GamesTotal = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: appName,
	Name:      "games_total",
	Help:      "Total number of games",
})

var GamesRunningTotal = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: appName,
	Name:      "games_running_total",
	Help:      "Total number of running games",
})

var ScoreTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: appName,
	Name:      "score_total",
	Help:      "Total score achieved",
})

var LinesClearedTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: appName,
	Name:      "lines_cleared_total",
	Help:      "Total lines cleared",
})

var PieceMovementsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: appName,
	Name:      "piece_movements_total",
	Help:      "Total count of piece movements by players",
}, []string{
	"piece", "rotation", "delta_x", "delta_y",
})

func IncreasePieceMovementsTotal(pieceToken piece.Token, dr piece.Rotation, dx int, dy int) {
	PieceMovementsTotal.With(prometheus.Labels{
		"piece":    pieceTokenNames[pieceToken],
		"rotation": strconv.Itoa(int(dr)),
		"delta_x":  strconv.Itoa(dx),
		"delta_y":  strconv.Itoa(dy),
	}).Inc()
}

var HardLocksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: appName,
	Name:      "hard_locks_total",
	Help:      "Total count of hard locks",
}, []string{
	"piece",
})

func IncreaseHardLocksTotal(pieceToken piece.Token) {
	HardLocksTotal.With(prometheus.Labels{
		"piece": pieceTokenNames[pieceToken],
	}).Inc()
}

var ItemActivationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: appName,
	Name:      "item_activations_total",
	Help:      "Total count of activated items",
}, []string{
	"item",
})

func IncreaseItemActivationsTotal(itemName string) {
	ItemActivationsTotal.With(prometheus.Labels{
		"item": itemName,
	}).Inc()
}

var BedrockSentTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: appName,
	Name:      "bedrock_sent_total",
	Help:      "Total count of bedrock sent to other players",
})
