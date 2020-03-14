// Code generated by sqlc. DO NOT EDIT.

package postgres

import (
	"github.com/byrdapp/timestamp/parser"
	"github.com/google/uuid"
)

type Booking struct {
	ID             uuid.UUID        `json:"id"`
	MediaID        string           `json:"media_id"`
	PhotographerID string           `json:"photographer_id"`
	Task           string           `json:"task"`
	Price          int32            `json:"price"`
	Credits        int32            `json:"credits"`
	Accepted       bool             `json:"accepted"`
	Completed      bool             `json:"completed"`
	DateStart      parser.Timestamp `json:"date_start"`
	DateEnd        parser.Timestamp `json:"date_end"`
	CreatedAt      parser.Timestamp `json:"created_at"`
	Lat            string           `json:"lat"`
	Lng            string           `json:"lng"`
}

type Profile struct {
	ID       uuid.UUID `json:"id"`
	UserID   string    `json:"user_id"`
	ProLevel int32     `json:"pro_level"`
}