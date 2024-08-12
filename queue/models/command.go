package models

type Command struct {
	Action string `json:"command"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	// temp
	GroupID       string
	ReceiptHandle *string
}

var AllowedActionsMap = map[string]bool{
	"addItem":     true,
	"deleteItem":  true,
	"getItem":     true,
	"getAllItems": true,
}
