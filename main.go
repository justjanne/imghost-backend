package main

import (
	"github.com/go-redis/redis"
	"encoding/json"
	"os"
	"gopkg.in/gographics/imagick.v2/imagick"
	"fmt"
	"path/filepath"
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

	errorChannel := make(chan error)

	for _, definition := range config.Sizes {
		go generateSize(errorChannel, config, image, definition)
	}

	errors := make([]string, 0)
	for i:= 0; i < len(config.Sizes); i++ {
		error := <- errorChannel
		if error != nil {
			errors = append(errors, error.Error())
		}
	}

	if len(errors) != 0 {
		returnResult(config, client, Result{
			Id:     image.Id,
			Success: false,
			Errors: errors,
		})
	} else {
		returnResult(config, client, Result{
			Id:      image.Id,
			Success: true,
		})
	}
}

func main() {
	config := Config{}

	json.Unmarshal([]byte(os.Getenv("IK8R_SIZES")), &config.Sizes)
	json.Unmarshal([]byte(os.Getenv("IK8R_QUALITY")), &config.Quality)
	config.SourceFolder = os.Getenv("IK8R_SOURCE_FOLDER")
	config.TargetFolder = os.Getenv("IK8R_TARGET_FOLDER")
	config.Redis.Address = os.Getenv("IK8R_REDIS_ADDRESS")
	config.Redis.Password = os.Getenv("IK8R_REDIS_PASSWORD")
	config.ImageQueue = os.Getenv("IK8R_REDIS_IMAGE_QUEUE")
	config.ResultChannel = os.Getenv("IK8R_REDIS_RESULT_CHANNEL")

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
