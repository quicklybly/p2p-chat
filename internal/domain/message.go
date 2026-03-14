package domain

import "time"

type Message struct {
	SenderID  string
	RoomName  string
	Text      string
	Timestamp time.Time
}
