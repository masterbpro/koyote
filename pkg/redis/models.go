package redis

import (
	"time"
)

type MessageType struct {
	ChatID   int64     `json:"chat_id"`
	ThreadID int       `json:"threadID,omitempty"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
}
