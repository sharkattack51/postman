package main

import (
	"fmt"
	"strings"

	"github.com/sharkattack51/golem"
)

//
// Secure
//

type SecureMessage struct {
	RawToken string `json:"token"`
	RawTkn   string `json:"tkn"`
}

func (m *SecureMessage) Token() string {
	if m.RawToken != "" {
		return m.RawToken
	} else {
		return m.RawTkn
	}
}

func NewSecureMessage(token string, tkn string) *SecureMessage {
	msg := &SecureMessage{
		RawToken: token,
		RawTkn:   tkn,
	}
	return msg
}

//
// Subscribe
//

type SubscribeMessage struct {
	RawChannel string `json:"channel"`
	RawCh      string `json:"ch"`
}

func (m *SubscribeMessage) Channel() string {
	if m.RawChannel != "" {
		return m.RawChannel
	} else {
		return m.RawCh
	}
}

//
// Publish
//

type PublishMessage struct {
	RawChannel   string `json:"channel"`
	RawCh        string `json:"ch"`
	RawMessage   string `json:"message"`
	RawMsg       string `json:"msg"`
	RawTag       string `json:"tag"`
	RawExtention string `json:"extention"`
	RawExt       string `json:"ext"`
}

func (m *PublishMessage) Channel() string {
	if m.RawChannel != "" {
		return m.RawChannel
	} else {
		return m.RawCh
	}
}

func (m *PublishMessage) Message() string {
	if m.RawMessage != "" {
		return m.RawMessage
	} else {
		return m.RawMsg
	}
}

func (m *PublishMessage) Tag() string {
	return m.RawTag
}

func (m *PublishMessage) Extention() string {
	if m.RawExtention != "" {
		return m.RawExtention
	} else {
		return m.RawExt
	}
}

func NewPublishMessage(channel string, ch string, message string, msg string, tag string, extention string, ext string) *PublishMessage {
	pmsg := &PublishMessage{
		RawChannel:   channel,
		RawCh:        ch,
		RawMessage:   message,
		RawMsg:       msg,
		RawTag:       tag,
		RawExtention: extention,
		RawExt:       ext,
	}
	return pmsg
}

type PublishSendMessage struct {
	Channel   string `json:"channel"`
	Message   string `json:"message"`
	Tag       string `json:"tag"`
	Extention string `json:"extention"`
}

func NewPublishSendMessage(channel string, message string, tag string, extention string) *PublishSendMessage {
	msg := &PublishSendMessage{
		Channel:   channel,
		Message:   message,
		Tag:       tag,
		Extention: extention,
	}
	return msg
}

func (msg *PublishMessage) BuildLogString() string {
	msgStr := ""
	if msg.Tag() != "" {
		if msg.Extention() != "" {
			msgStr = fmt.Sprintf("%s/%s/%s", msg.Message(), msg.Tag(), msg.Extention())
		} else {
			msgStr = fmt.Sprintf("%s/%s", msg.Message(), msg.Tag())
		}
	} else {
		if msg.Extention() != "" {
			msgStr = fmt.Sprintf("%s/%s", msg.Message(), msg.Extention())
		} else {
			msgStr = msg.Message()
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

//
// Store
//

type StoreMessage struct {
	RawCommand string `json:"command"`
	RawCmd     string `json:"cmd"`
	RawKey     string `json:"key"`
	RawValue   string `json:"value"`
	RawVal     string `json:"val"`
}

func (m *StoreMessage) Command() string {
	if m.RawCommand != "" {
		return m.RawCommand
	} else {
		return m.RawCmd
	}
}

func (m *StoreMessage) Key() string {
	return m.RawKey
}

func (m *StoreMessage) Value() string {
	if m.RawValue != "" {
		return m.RawValue
	} else {
		return m.RawVal
	}
}

func NewStoreMessage(command string, cmd string, key string, value string, val string) *StoreMessage {
	msg := &StoreMessage{
		RawCommand: command,
		RawCmd:     cmd,
		RawKey:     key,
		RawValue:   value,
		RawVal:     val,
	}
	return msg
}

//
// Result
//

type ResultMessage struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

func NewResultMessage(result string, err string) *ResultMessage {
	msg := &ResultMessage{
		Result: result,
		Error:  err,
	}
	return msg
}
