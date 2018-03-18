package main

import (
	"github.com/go-redis/redis"
	"encoding/json"
	"os"
	"gopkg.in/gographics/imagick.v2/imagick"
	"fmt"
	"path/filepath"
	"time"
	"net/http"
)

func returnResult(config *Config, client *redis.Client, result Result) {
	raw, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	client.Publish(config.ResultChannel, string(raw))
}

func generateSize(errorChannel chan error, wand *imagick.MagickWand, wandLinear *imagick.MagickWand, config *Config, image Image, definition SizeDefinition) {
	errorChannel <- resize(
		wand,
		wandLinear,
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

	wand := imagick.NewMagickWand()

	err := wand.ReadImage(filepath.Join(config.SourceFolder, image.Id))
	if err != nil {
		panic(err)
	}

	wandLinear := wand.Clone()
	err = wandLinear.TransformImageColorspace(imagick.COLORSPACE_RGB)
	if err != nil {
		panic(err)
	}

	for _, definition := range config.Sizes {
		go generateSize(errorChannel, wand, wandLinear, config, image, definition)
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

	wand.Clear()
	wandLinear.Clear()
}

func main() {
	go func() {
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
	}()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
