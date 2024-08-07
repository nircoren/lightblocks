package models

type Message struct {
	Command string `json:"command"`
	Key     string `json:"key,omitempty"`
	Value   string `json:"value,omitempty"`
}
