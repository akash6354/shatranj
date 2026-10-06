package cache

const keyPrefix = "shatranj:"

// UserKey returns the Redis key for a user record.
func UserKey(userID string) string {
	return keyPrefix + "user:" + userID
}

// GameKey returns the Redis key for a game state.
func GameKey(gameID string) string {
	return keyPrefix + "game:" + gameID
}
