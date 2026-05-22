package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var downloadedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "bot",
	Subsystem: "downloader",
	Name:      "downloaded_total",
	Help:      "Total number of downloaded videos",
}, []string{"service"})
