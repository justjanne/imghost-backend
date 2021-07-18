package main

import (
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/gographics/imagick.v2/imagick"
	"net/http"
)

var queueGauge = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "imghost_waiting_images",
		Help: "The number of waiting image events",
	},
	[]string{"state"},
)
var queueGaugeQueued = queueGauge.WithLabelValues("queued")
var queueGaugeInProgress = queueGauge.WithLabelValues("inProgress")

var imageCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "imghost_processed_images_total",
		Help: "The total number of successfully processed image events",
	},
	[]string{"result"},
)
var imageCounterSuccess = imageCounter.WithLabelValues("success")
var imageCounterFailure = imageCounter.WithLabelValues("failure")

var imageProcessDuration = promauto.NewCounter(prometheus.CounterOpts{
	Name: "imghost_process_duration",
	Help: "The total amount of time spent processing images",
})

func main() {
	config := NewConfigFromEnv()

	imagick.Initialize()
	defer imagick.Terminate()

	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
	})

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	go serveQueue(client, config.ImageQueue, func(value string) {
		queueGaugeQueued.Dec()
		ProcessImage(&config, client, value)
	})
	if err := http.ListenAndServe(":2112", nil); err != nil {
		panic(err)
	}
}
