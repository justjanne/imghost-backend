package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
)

func serveQueue(client *redis.Client, queue string, function func(value string)) {
	for {
		element := client.BLPop(0, fmt.Sprintf("queue:%s", queue))
		if len(element.Val()) == 2 {
			value := element.Val()[1]
			queueGauge.Inc()
			go function(value)
		}
	}
}

func returnResult(config *Config, client *redis.Client, result Result) {
	raw, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	client.Publish(config.ResultChannel, string(raw))
}
