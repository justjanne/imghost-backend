package main

import (
	"github.com/go-redis/redis"
	"encoding/json"
	"os"
	"gopkg.in/gographics/imagick.v2/imagick"
	"fmt"
	"path/filepath"
	"time"
)

func returnResult(config *Config, client *redis.Client, result Result) {
	raw, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	client.Publish(config.ResultChannel, string(raw))
}

func generateSize(errorChannel chan error, config *Config, image Image, definition SizeDefinition) {
	errorChannel <- resize(
		filepath.Join(config.SourceFolder, image.Id),
		definition.Size,
		config.Quality,
		filepath.Join(config.TargetFolder, fmt.Sprintf("%s%s", image.Id, definition.Suffix)),
	)
}

func processImage(config *Config, client *redis.Client, value string) {
	image := Image{}
	if err := json.Unmarshal([]byte(value), &image); err != nil {
		panic(err)
	}
	fmt.Printf("Received task %s at %d\n", image.Id, time.Now().Unix())

	errorChannel := make(chan error)

	for _, definition := range config.Sizes {
		go generateSize(errorChannel, config, image, definition)
	}

	errors := make([]string, 0)
	for i := 0; i < len(config.Sizes); i++ {
		err := <-errorChannel
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	os.Remove(filepath.Join(config.SourceFolder, image.Id))

	fmt.Printf("Finished task %s at %d\n", image.Id, time.Now().Unix())

	if len(errors) != 0 {
		returnResult(config, client, Result{
			Id:      image.Id,
			Success: false,
			Errors:  errors,
		})
	} else {
		returnResult(config, client, Result{
			Id:      image.Id,
			Success: true,
		})
	}
}

func main() {
	config := NewConfigFromEnv()

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
}
