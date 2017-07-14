package main

import (
	"log"
	"encoding/json"
)

type Command interface {
	Handle(*Client)
}

type TextMessageCommand struct {
	Message string `json:"message"`
}

func (cmd *TextMessageCommand) Handle(c *Client) {
	socketmanager.BroadcastTextMessage(cmd.Message, c.ID)
}

func ParseCommand(cmdName []byte, data []byte) Command {
	var command Command

	switch string(cmdName) {
	case "textmessage":
		command = &TextMessageCommand{}
	default:
		log.Println("Command", cmdName, "does not exist")
	}

	err := json.Unmarshal(data, &command)

	if err != nil {
		panic(err)
	}

	return command
}