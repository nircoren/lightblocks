package models

import (
	"fmt"
)

type CommandBase struct {
	Action string `json:"action"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
}

type Command struct {
	CommandBase
	GroupID       string
	ReceiptHandle *string
}

func (c *Command) Validate() error {
	switch c.Action {
	case "getAllItems":
		if c.Key != "" || c.Value != "" {
			return fmt.Errorf("invalid command: getAllItems should not have key or value")
		}
	case "addItem":
		if c.Key == "" || c.Value == "" {
			return fmt.Errorf("invalid command: addItem must have both key and value")
		}
	case "deleteItem", "getItem":
		if c.Key == "" {
			return fmt.Errorf("invalid command: %s must have a key", c.Action)
		}
		if c.Value != "" {
			return fmt.Errorf("invalid command: %s should not have a value", c.Action)
		}
	default:
		return fmt.Errorf("invalid command: %s is not a recognized action", c.Action)
	}
	return nil
}
