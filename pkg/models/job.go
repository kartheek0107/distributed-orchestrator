package models

type Job struct {
	ID        string `json:"id"`
	Command   string `json:"command"`
	Timeout   int    `json:"timeout"` // in seconds
	Status    string `json:"status"`
	Priority  int    `json:"priority"`
	CreatedAt int64  `json:"created_at"`
}


