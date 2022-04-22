package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"path/filepath"
	"time"
)

func trackTime(start time.Time, function func(time.Duration)) {
	function(time.Now().UTC().Sub(start.UTC()))
}

func ProcessImage(ctx context.Context, config *Config, client *redis.Client, value string) {
	queueGaugeInProgress.Inc()
	defer queueGaugeInProgress.Dec()
	defer trackTime(time.Now(), func(duration time.Duration) {
		imageProcessDuration.Add(float64(duration.Milliseconds()))
	})

	image := Image{}
	if err := json.Unmarshal([]byte(value), &image); err != nil {
		fmt.Printf("Could not unmarshal task %s\n", value)
		return
	}

	errors := ResizeImage(config, image.Id)
	_ = os.Remove(filepath.Join(config.SourceFolder, image.Id))

	errorMessages := make([]string, len(errors))
	for i, err := range errors {
		errorMessages[i] = err.Error()
	}

	if len(errors) != 0 {
		imageCounterFailure.Inc()
		returnResult(ctx, config, client, Result{
			Id:      image.Id,
			Success: false,
			Errors:  errorMessages,
		})
	} else {
		imageCounterSuccess.Inc()
		returnResult(ctx, config, client, Result{
			Id:      image.Id,
			Success: true,
		})
	}
}
