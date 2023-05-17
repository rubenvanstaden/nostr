package core

type EventId string

type Event struct {
	Id      EventId `json:"id"`
	Content string  `json:"content"`
	Kind    uint8   `json:"kind,string"`
}
