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

type RaidEvent struct {
	Subscription helix.EventSubSubscription      	`json:"subscription"`
	Event        helix.EventSubChannelRaidEvent 	`json:"event"`
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
	case helix.EventSubTypeChannelRaid:
		var parsedBody RaidEvent
		ParseErr := json.Unmarshal(body, &parsedBody)
		if ParseErr != nil {
			log.Printf("Could not read body: %s\n", ParseErr)
			return &HTTPError{
				HTTPStatus: http.StatusBadRequest,
				Message:    "Couldn't parse follow event as JSON object.",
			}
		}
		bot.handleRaid(parsedBody);
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

func (bot TwitchBot) handleRaid(event RaidEvent) {
	bot.shoutout(
		event.Event.FromBroadcasterUserID,
		fmt.Sprintf("Please take a moment to check out the amazing %s at https://twitch.tv/%s !", event.Event.FromBroadcasterUserName, event.Event.FromBroadcasterUserLogin),
	)
}

func handleFollow(event FollowEvent) {
	log.Println("Follow event handling not implemented yet !")
}

func (bot TwitchBot) handleChatMessage(event ChatMessageEvent) {
	if !slices.Contains(bot.cfg.ShouthoutList, event.Event.ChatterUserName) {
		return
	}
	if _, present := bot.shoutouts[event.Event.ChatterUserID]; !present {
		bot.shoutout(
			event.Event.ChatterUserID,
			fmt.Sprintf("Please take a moment to check out the amazing %s at https://twitch.tv/%s !", event.Event.ChatterUserName, event.Event.ChatterUserLogin),
		)
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
		Message:       bot.cfg.JoinMessage,
	})
	if err != nil {
		log.Printf("Couldn't send shoutout message: %s", err)
		return
	}
}

func (bot TwitchBot) shoutout(userID, message string) {
	var err error
	shoutoutAnswer, err := bot.userClient.SendShoutout(&helix.SendShoutoutParams{
		FromBroadcasterID: bot.cfg.BroadcasterID,
		ToBroadcasterID:   userID,
		ModeratorID:       bot.cfg.BotUserID,
	})
	log.Printf("answer: %v", shoutoutAnswer)
	if err != nil || (shoutoutAnswer.StatusCode >= 400) {
		log.Printf("Couldn't send shoutout for user %s, twitch answered: %#v, err: %v", userID, shoutoutAnswer, err);
		return
	}
	time.Sleep(1 * time.Second);
	announcementAnswer, err := bot.userClient.SendChatAnnouncement(&helix.SendChatAnnouncementParams{
		BroadcasterID: 		bot.cfg.BroadcasterID,
		ModeratorID:      bot.cfg.BotUserID,
		Message:       		message,
	})
	if err != nil || (announcementAnswer.StatusCode >= 400) {
		log.Printf("Couldn't send announccement shoutout for user %s, twitch answered: %#v, err: %v", userID, shoutoutAnswer, err);
		return;
	}
}
