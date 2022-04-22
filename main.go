package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/gographics/imagick.v3/imagick"
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

var imageProcessDuration = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "imghost_process_duration",
	Help: "The amount of time spent processing images",
}, []string{"task"})
var imageProcessDurationRead = imageProcessDuration.WithLabelValues("read")
var imageProcessDurationClone = imageProcessDuration.WithLabelValues("clone")
var imageProcessDurationCrop = imageProcessDuration.WithLabelValues("crop")
var imageProcessDurationResize = imageProcessDuration.WithLabelValues("resize")
var imageProcessDurationWrite = imageProcessDuration.WithLabelValues("write")

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

	go serveQueue(context.Background(), client, config.ImageQueue, func(ctx context.Context, value string) {
		queueGaugeQueued.Dec()
		ProcessImage(ctx, &config, client, value)
	})
	if err := http.ListenAndServe(":2112", nil); err != nil {
		panic(err)
	}
}
