package models

type Meeting struct {
	ID        string `json:"id"`
	ZoomID    string `json:"zoom_id"`
	Topic     string `json:"topic"`
	Agenda    string `json:"agenda"`
	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	JoinURL   string `json:"join_url"`
	StartURL  string `json:"start_url"`
}
