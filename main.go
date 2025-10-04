package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"

	"nodebot/internal/db"
	"nodebot/internal/twitch"
)

type Conf struct {
	TwitchConfig 		twitch.TwitchConfig `yaml:"twitchConfig"`
	DatabaseConfig 	db.DatabaseConfig		`yaml:"database"`
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
	TwitchBot := twitch.NewTwitchBot(config.TwitchConfig, config.DatabaseConfig)
	if TwitchBot == nil {
		log.Fatal("Couldn't initialize twitch bot.")
	}

	wg.Go(func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/twitch/eventsub", TwitchBot.TwitchWebHookHandler)
		err := http.ListenAndServe(":3333", nil)
		if err != nil {
			log.Printf("Couldn't start HTTP server: %s \n", err)
			os.Exit(1)
		}
		log.Print("Server started")
	})
	// start subscription manager
	TwitchBot.SubscribeToEvents()
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
