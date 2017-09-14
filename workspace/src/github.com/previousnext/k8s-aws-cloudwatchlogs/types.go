package main

import "time"

// Message is for each line in the file.
type Message struct {
	Log    string    `json:"log"`
	Stream string    `json:"stream"`
	Time   time.Time `json:"time"`
}
