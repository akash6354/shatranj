package realtime

import (
	"encoding/json"
	"testing"
)

func TestChatHubPresenceAndRoomBroadcast(t *testing.T) {
	hub := NewHub(nil)
	client := &ChatClient{
		send:   make(chan []byte, 1),
		userID: "user",
		roomID: "room",
	}

	hub.registerChat(client)
	online, count := hub.Presence("user")
	if !online || count != 1 {
		t.Fatalf("Presence() = (%t, %d), want (true, 1)", online, count)
	}

	hub.PublishRoom("room", Event{Type: EventPresenceUpdated, UserID: "user", Online: true, Count: count})
	var event Event
	if err := json.Unmarshal(<-client.send, &event); err != nil {
		t.Fatalf("decode published event: %v", err)
	}
	if event.Type != EventPresenceUpdated || event.UserID != "user" || !event.Online || event.Count != 1 {
		t.Fatalf("published event = %+v", event)
	}

	hub.unregisterChat(client)
	online, count = hub.Presence("user")
	if online || count != 0 {
		t.Fatalf("Presence() after unregister = (%t, %d), want (false, 0)", online, count)
	}
}
