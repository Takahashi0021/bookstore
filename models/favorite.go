package models

import "time"

type Favorite struct {
	UserID    int       `json:"user_id"`
	BookID    int       `json:"book_id"`
	CreatedAt time.Time `json:"created_at"`
}

type FavoriteRequest struct {
	BookID int `json:"book_id"`
}
