package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"runtime"
	"sync"

	"nodebot/internal/twitch"
)

type Conf struct {
	TwitchConfig twitch.TwitchConfig `yaml:"twitchConfig"`
}

func main() {
	// read config
	path, exists := os.LookupEnv("BOT_CONFIGPATH")
	if !exists {
		path = "/etc/nodebot/config.yaml"
	}
	log.Printf("Reading config from: %s \n", path)
	config, configErr := readConfig(path)
	if configErr != nil {
		log.Printf("There was an error reading config: %s \n", configErr)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	// start twitch webhook handler
	TwitchBot := twitch.NewTwitchBot(config.TwitchConfig)
	if TwitchBot == nil {
		log.Fatal("Couldn't initialize twitch bot.")
	}
	wg.Go(func() {
		TwitchBot.StartWebhook()
	})
	// start subscription manager
	TwitchBot.SubscribeToEvents()
	fmt.Printf("go routines: %v \n", runtime.NumGoroutine())
	wg.Wait()
}

func readConfig(configPath string) (*Conf, error) {
	f, ReadErr := os.ReadFile(configPath)
	if ReadErr != nil {
		return nil, ReadErr
	}

	var parsedConfig Conf
	ParseErr := yaml.Unmarshal(f, &parsedConfig)
	if ParseErr != nil {
		return nil, ParseErr
	}
	return &parsedConfig, nil
}
