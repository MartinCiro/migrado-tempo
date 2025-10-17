package domain

import "time"

type Email struct {
	ID      uint32
	From    string
	Subject string
	Date    time.Time
	Read    bool
}

type EmailCriteria struct {
	UnreadOnly bool
}
