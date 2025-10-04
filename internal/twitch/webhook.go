package twitch

import (
	"context"
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
	"github.com/jackc/pgx/v5"

	"nodebot/internal/db"
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
	dbClient *pgx.Conn
	shoutouts map[string]time.Time
}

func NewTwitchBot(twitchConfig TwitchConfig, dbConfig db.DatabaseConfig) *TwitchBot {
var err error;
appClient, err := helix.NewClient(&helix.Options{
		ClientID:       twitchConfig.ClientID,
		ClientSecret:   twitchConfig.ClientSecret,
		AppAccessToken: twitchConfig.AppAccesToken,
	})
	if err != nil {
		log.Printf("Couldn't initialize twitch bot's API client: %s", err);
		return nil
	}
	userClient, err := helix.NewClient(&helix.Options{
		ClientID:       twitchConfig.ClientID,
		ClientSecret:   twitchConfig.ClientSecret,
		AppAccessToken: twitchConfig.UserAccessToken,
	})
	if err != nil {
		log.Printf("Couldn't initialize twitch bot's API client: %s", err);
		return nil
	}
	pgConn, err := pgx.Connect(context.Background(), db.GenerateConnString(dbConfig))
	if err != nil {
		log.Printf("Unable to connect to database: %s\n", err);
		return nil
	}

	return &TwitchBot {
		cfg: twitchConfig,
		appClient: appClient,
		userClient: userClient,
		dbClient: pgConn,
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

	// Verify this is not a duplicate event
	commandTag, err := bot.dbClient.Exec(context.Background(), "SELECT id FROM twitch_notifications WHERE id=$1", r.Header.Get("Twitch-Eventsub-Message-Id"))
	if err != nil {
	  log.Printf("Error checking db for processed events: %s", err)
	}
	if commandTag.RowsAffected() == 1 {
		log.Printf("Message already treated")
		w.WriteHeader(http.StatusOK)
		return
	}

	//Handle message type
	switch r.Header.Get("Twitch-Eventsub-Message-Type") {
		case "notification":
			err := bot.processEvent(r.Header.Get("Twitch-Eventsub-Subscription-Type"), rawBody)
			bot.storeEventID(r.Header.Get("Twitch-Eventsub-Message-Id"))
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

func (bot *TwitchBot) storeEventID(eventID string) {
	err := pgx.BeginFunc(context.Background(), bot.dbClient, func(tx pgx.Tx) error {
    _, err := tx.Exec(context.Background(), "INSERT INTO twitch_notifications VALUES ($1, transaction_timestamp())", eventID)
    return err
	})
	if err != nil {
    log.Printf("Couldn't store event ID: %s, reason: %v\n", eventID, err);
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
