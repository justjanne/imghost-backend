package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
)

func serveQueue(
	ctx context.Context,
	client *redis.Client,
	queue string,
	function func(ctx context.Context, value string),
) {
	for {
		element := client.BLPop(ctx, 0, fmt.Sprintf("queue:%s", queue))
		if len(element.Val()) == 2 {
			value := element.Val()[1]
			queueGaugeQueued.Inc()
			go function(ctx, value)
		}
	}
}

func returnResult(ctx context.Context, config *Config, client *redis.Client, result Result) {
	raw, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	client.Publish(ctx, config.ResultChannel, string(raw))
}
