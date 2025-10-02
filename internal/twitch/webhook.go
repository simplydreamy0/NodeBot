package twitch

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nicklaw5/helix/v2"
)

type HTTPError struct {
	error
	HTTPStatus int
	Message    string
}

type TwitchConfig struct {
	ClientID      		string        `yaml:"clientID"`
	ClientSecret  		string        `yaml:"clientSecret"`
	AppAccesToken 		string        `yaml:"appAccessToken"`
	UserAccessToken 	string      	`yaml:"userAccessToken"`
	UserRefreshToken 	string    		`yaml:"userRefreshToken"`
	BotUserID     		string        `yaml:"botUserID"`
	BroadcasterID 		string        `yaml:"broadcasterID"`
	WebhookConfig 		WebhookConfig `yaml:"webhookConfig"`
	ShouthoutList 		[]string			`yaml:"shouthoutList"`
}

type WebhookConfig struct {
	Url    string `yaml:"url"`
	Secret string `yaml:"secret"`
}

type TwitchBot struct {
	cfg TwitchConfig
	appClient *helix.Client
	userClient *helix.Client
	shoutouts map[string]time.Time
}

func NewTwitchBot(config TwitchConfig) *TwitchBot {
	var err error;
	appClient, err := helix.NewClient(&helix.Options{
		ClientID:       config.ClientID,
		ClientSecret:   config.ClientSecret,
		AppAccessToken: config.AppAccesToken,
	})
	if err != nil {
		log.Printf("Couldn't initialize twitch bot's API client: %s", err);
		return nil
	}
	userClient, err := helix.NewClient(&helix.Options{
		ClientID:       config.ClientID,
		ClientSecret:   config.ClientSecret,
		AppAccessToken: config.UserAccessToken,
	})
	if err != nil {
		log.Printf("Couldn't initialize twitch bot's API client: %s", err);
		return nil
	}
	return &TwitchBot {
		cfg: config,
		appClient: appClient,
		userClient: userClient,
		shoutouts: make(map[string]time.Time),
	}
}

func (bot *TwitchBot) TwitchWebHookHandler(w http.ResponseWriter, r *http.Request) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error parsing raw body: %v \n", err)
		http.Error(w, "Couldn't parse body", http.StatusBadRequest)
	}

	// Verify the message comes from Twitch
	msg := r.Header.Get("Twitch-Eventsub-Message-Id") + r.Header.Get("Twitch-Eventsub-Message-Timestamp") + string(rawBody)
	sig := strings.ReplaceAll(r.Header.Get("Twitch-Eventsub-Message-Signature"), "sha256=", "")
	if equal, err := verifyHMAC(msg, bot.cfg.WebhookConfig.Secret, sig); !equal {
		log.Printf("Error verifying signature: %v", err)
		http.Error(w, "Couldn't verify signature", http.StatusForbidden)
	}

	//Handle message type
	switch r.Header.Get("Twitch-Eventsub-Message-Type") {
		case "notification":
			err := bot.processEvent(r.Header.Get("Twitch-Eventsub-Subscription-Type"), rawBody)
			if err != nil {
				http.Error(w, err.Message, err.HTTPStatus)
				return
			}
			w.WriteHeader(http.StatusOK)
		case "webhook_callback_verification":
			challenge, err := processChallenge(rawBody)
			if err != nil {
				http.Error(w, err.Message, err.HTTPStatus)
				return
			}
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(*challenge)); err != nil {
				log.Println("Error writing challenge back")
			}
		case "revocation":
			//TODO
			//processRevocation();
			panic("Revocation handling not implemented")
		default:
			fmt.Printf("Couldn't process message type: %s \n", r.Header.Get("Twitch-Eventsub-Message-Type"))
	}
}

// Verify twitchSig based on msg & webhookSecret
func verifyHMAC(msg, webhookSecret, twitchSig string) (bool, error) {
	secret := []byte(webhookSecret)
	sig, err := hex.DecodeString(twitchSig)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(msg))
	return hmac.Equal(sig, mac.Sum(nil)), nil
}
