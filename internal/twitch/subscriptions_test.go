package twitch

import (
	"testing"
)


func TestProcessChallenge(t *testing.T) {
	body := []byte(`{"challenge":"98d60329-c81c-8d4f-49e2-5b5d495f2df7","subscription":{"id":"55ce0852-fa1c-59c8-9df3-305803f2d4b1","status":"webhook_callback_verification_pending","type":"stream.online","version":"1","condition":{"broadcaster_user_id":"2809544"},"transport":{"method":"webhook","callback":"http://localhost:3333"},"created_at":"2025-09-23T18:50:31.906035Z","cost":0}}`)
	wants := "98d60329-c81c-8d4f-49e2-5b5d495f2df7"

	if challenge, _ := processChallenge([]byte(body)); *challenge != wants {
		t.Errorf("Got wrong challenge, got: %s, wants, %s", *challenge, wants);
	}
}
