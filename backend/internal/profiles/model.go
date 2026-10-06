package profiles

import "time"

type RatingsSummary struct {
	Bullet    int `json:"bullet"`
	Blitz     int `json:"blitz"`
	Rapid     int `json:"rapid"`
	Classical int `json:"classical"`
}

type Profile struct {
	UserID      string         `json:"user_id"`
	Username    string         `json:"username"`
	DisplayName string         `json:"display_name"`
	AvatarURL   string         `json:"avatar_url,omitempty"`
	Country     string         `json:"country,omitempty"`
	Bio         string         `json:"bio,omitempty"`
	Ratings     RatingsSummary `json:"ratings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type UpdateInput struct {
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Country     *string `json:"country"`
	Bio         *string `json:"bio"`
}
