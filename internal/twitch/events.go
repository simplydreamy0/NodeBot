package twitch

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/nicklaw5/helix/v2"
)

type FollowEvent struct {
	Subscription helix.EventSubSubscription       `json:"subscription"`
	Event        helix.EventSubChannelFollowEvent `json:"event"`
}

type ChatMessageEvent struct {
	Subscription helix.EventSubSubscription            `json:"subscription"`
	Event        helix.EventSubChannelChatMessageEvent `json:"event"`
}

type SteamOnlineEvent struct {
	Subscription helix.EventSubSubscription      `json:"subscription"`
	Event        helix.EventSubStreamOnlineEvent `json:"event"`
}

func (bot TwitchBot) processEvent(subscriptionType string, body []byte) *HTTPError {
	switch subscriptionType {
	case helix.EventSubTypeChannelFollow:
		var parsedBody FollowEvent
		ParseErr := json.Unmarshal(body, &parsedBody)
		if ParseErr != nil {
			log.Printf("Could not read body: %s\n", ParseErr)
			return &HTTPError{
				HTTPStatus: http.StatusBadRequest,
				Message:    "Couldn't parse follow event as JSON object.",
			}
		}
		handleFollow(parsedBody)
		return nil
	case helix.EventSubTypeChannelChatMessage:
		var parsedBody ChatMessageEvent
		ParseErr := json.Unmarshal(body, &parsedBody)
		if ParseErr != nil {
			log.Printf("Could not read body: %s\n", ParseErr)
			return &HTTPError{
				HTTPStatus: http.StatusBadRequest,
				Message:    "Couldn't parse chat message event as JSON object.",
			}
		}
		bot.handleChatMessage(parsedBody)
		return nil
	case helix.EventSubTypeStreamOnline:
		var parsedBody SteamOnlineEvent
		ParseErr := json.Unmarshal(body, &parsedBody)
		if ParseErr != nil {
			log.Printf("Could not read body: %s\n", ParseErr)
			return &HTTPError{
				HTTPStatus: http.StatusBadRequest,
				Message:    "Couldn't parse stream online event as JSON object.",
			}
		}
		bot.handleStreamOnline(parsedBody)
		return nil
	default:
		fmt.Printf("Couldn't parse event type: %s \n", subscriptionType)
		return &HTTPError{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintf("Couldn't parse event type: %s \n", subscriptionType),
		}
	}
}

func handleFollow(event FollowEvent) {
	log.Println("Follow event handling not implemented yet !")
}

func (bot TwitchBot) handleChatMessage(event ChatMessageEvent) {
	if !slices.Contains(bot.cfg.ShouthoutList, event.Event.ChatterUserName) {
		return
	}
	if _, present := bot.shoutouts[event.Event.ChatterUserID]; !present {
		bot.shoutout(event.Event.ChatterUserID, event.Event.ChatterUserLogin)
		bot.shoutouts[event.Event.ChatterUserID] = time.Now()
	} else {
		lastShoutout := bot.shoutouts[event.Event.ChatterUserID]
		if time.Since(lastShoutout).Hours() >= 3 {
			bot.shoutout(event.Event.ChatterUserID, event.Event.ChatterUserLogin)
			bot.shoutouts[event.Event.ChatterUserID] = time.Now()
		}
	}
}

func (bot TwitchBot) handleStreamOnline(event SteamOnlineEvent) {
	var err error
	_, err = bot.appClient.SendChatMessage(&helix.SendChatMessageParams{
		BroadcasterID: bot.cfg.BroadcasterID,
		SenderID:      bot.cfg.BotUserID,
		Message:       "NodeBot joined the chat ! catArriveDestoy",
	})
	if err != nil {
		log.Printf("Couldn't send shoutout message: %s", err)
		return
	}
}

func (bot TwitchBot) shoutout(userID, username string) {
	var err error
	answer, err := bot.userClient.SendShoutout(&helix.SendShoutoutParams{
		FromBroadcasterID: bot.cfg.BroadcasterID,
		ToBroadcasterID:   userID,
		ModeratorID:       bot.cfg.BotUserID,
	})
	log.Printf("answer: %v", answer)
	if err != nil || (answer.StatusCode < 200 && answer.StatusCode >= 300) {
		log.Printf("Couldn't send shoutout for user %s: %s", userID, err)
		return
	}
	_, err = bot.appClient.SendChatAnnouncement(&helix.SendChatAnnouncementParams{
		BroadcasterID: 		bot.cfg.BroadcasterID,
		ModeratorID:      bot.cfg.BotUserID,
		Message:       		fmt.Sprintf("Please take a moment to check out the amazing %s at https://twitch.tv/%s !", username, username),
	})
	if err != nil {
		log.Printf("Couldn't send shoutout message: %s", err)
		return;
	}
}
