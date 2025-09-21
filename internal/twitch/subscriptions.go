package twitch

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nicklaw5/helix/v2"
)

type Challenge struct {
	Challenge    string                     `json:"challenge"`
	Subscription helix.EventSubSubscription `json:"subscription"`
}

func processChallenge(body []byte) (*string, *HTTPError) {
	var parsedBody Challenge
	ParseErr := json.Unmarshal(body, &parsedBody)
	if ParseErr != nil {
		log.Printf("Could parse body as JSON: %s\n", ParseErr)
		return nil, &HTTPError{
			HTTPStatus: http.StatusBadRequest,
			Message:    "Couldn't parse challenge as JSON object",
		}
	}

	return &parsedBody.Challenge, nil
}

func (bot TwitchBot) SubscribeToEvents() {
	var err error;
	// Subscribe to chat message
	_, err = bot.appClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelChatMessage,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: bot.cfg.BroadcasterID,
			UserID:            bot.cfg.BotUserID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: bot.cfg.WebhookConfig.Url,
			Secret:   bot.cfg.WebhookConfig.Secret,
		},
	})
	if err != nil {
		log.Printf("%v \n", err)
	}

	// Subscribe to stream online
	_, err = bot.appClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeStreamOnline,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: bot.cfg.BroadcasterID,
			UserID:            bot.cfg.BotUserID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: bot.cfg.WebhookConfig.Url,
			Secret:   bot.cfg.WebhookConfig.Secret,
		},
	})
	if err != nil {
		log.Printf("%v \n", err)
	}
}
