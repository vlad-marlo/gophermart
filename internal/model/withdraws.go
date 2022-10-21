package model

import "time"

type Withdraw struct {
	Order             int       `json:"order,string"`
	Sum               int       `json:"sum"`
	ProcessedAt       time.Time `json:"-"`
	ProcessedAtString string    `json:"processed_at,omitempty"`
}

func (w *Withdraw) ToRepresentation() {
	w.ProcessedAtString = w.ProcessedAt.Format(time.RFC3339)
}
