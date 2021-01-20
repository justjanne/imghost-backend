package main

import (
	"encoding/json"
	"fmt"
	"git.kuschku.de/justjanne/imghost/lib"
	"github.com/go-redis/redis"
	"gopkg.in/gographics/imagick.v2/imagick"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func returnResult(config *lib.Config, client *redis.Client, result lib.Result) {
	raw, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	client.Publish(config.ResultChannel, string(raw))
}

func processImage(config *lib.Config, client *redis.Client, value string) {
	image := lib.Image{}
	if err := json.Unmarshal([]byte(value), &image); err != nil {
		panic(err)
	}
	fmt.Printf("Received task %s at %d\n", image.Id, time.Now().Unix())

	errors := lib.ResizeImage(config, image.Id)
	_ = os.Remove(filepath.Join(config.SourceFolder, image.Id))

	errorMessages := make([]string, len(errors))
	for i, err := range errors {
		errorMessages[i] = err.Error()
	}

	fmt.Printf("Finished task %s at %d\n", image.Id, time.Now().Unix())

	if len(errors) != 0 {
		returnResult(config, client, lib.Result{
			Id:      image.Id,
			Success: false,
			Errors:  errorMessages,
		})
	} else {
		returnResult(config, client, lib.Result{
			Id:      image.Id,
			Success: true,
		})
	}
}

func main() {
	go func() {
		config := lib.NewConfigFromEnv()

		imagick.Initialize()
		defer imagick.Terminate()

		client := redis.NewClient(&redis.Options{
			Addr:     config.Redis.Address,
			Password: config.Redis.Password,
		})

		for {
			element := client.BLPop(0, fmt.Sprintf("queue:%s", config.ImageQueue))
			if len(element.Val()) == 2 {
				value := element.Val()[1]
				go processImage(&config, client, value)
			}
		}
	}()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
