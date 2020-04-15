package user

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	TLD       string    `json:"tld"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
