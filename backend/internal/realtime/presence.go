package realtime

type Presence struct {
	UserID string `json:"user_id"`
	Online bool   `json:"online"`
	Count  int    `json:"connections"`
}

func (h *Hub) GetPresence(userID string) Presence {
	online, count := h.Presence(userID)
	return Presence{UserID: userID, Online: online, Count: count}
}
