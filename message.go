package main

import (
	"fmt"
	"strings"

	"github.com/sharkattack51/golem"
)

//
// Subscribe
//

type SubscribeMessage struct {
	Channel string `json:"channel"`
}

//
// Publish
//

type PublishMessage struct {
	Channel   string `json:"channel"`
	Message   string `json:"message"`
	Tag       string `json:"tag"`
	Extention string `json:"extention"`
}

func (msg *PublishMessage) BuildLogString() string {
	msgStr := ""
	if msg.Tag != "" {
		if msg.Extention != "" {
			msgStr = fmt.Sprintf("%s:%s:%s", msg.Message, msg.Tag, msg.Extention)
		} else {
			msgStr = fmt.Sprintf("%s:%s", msg.Message, msg.Tag)
		}
	} else {
		if msg.Extention != "" {
			msgStr = fmt.Sprintf("%s:%s", msg.Message, msg.Extention)
		} else {
			msgStr = msg.Message
		}
	}

	return msgStr
}

//
// Status
//

type StatusMessage struct {
	Version  string              `json:"version"`
	Channels map[string][]string `json:"channels"`
}

func NewStatusMessage(rm *golem.RoomManager) *StatusMessage {
	channels := make(map[string][]string)

	if rm != nil {
		for _, ri := range rm.GetRoomInfos() {
			addrs := []string{}
			for _, c := range ri.Room.GetMembers() {
				addr := strings.Split(c.GetSocket().RemoteAddr().String(), ":")[0]
				addrs = append(addrs, addr)
			}
			channels[ri.Topic] = addrs
		}
	}

	msg := &StatusMessage{
		Version:  VERSION,
		Channels: channels,
	}
	return msg
}
