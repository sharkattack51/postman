package main

import (
	"fmt"

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
	RawChannel    string `json:"channel"`
	RawCh         string `json:"ch"`
	RawClientInfo string `json:"client_info"`
	RawCi         string `json:"ci"`
}

func (m *SubscribeMessage) Channel() string {
	if m.RawChannel != "" {
		return m.RawChannel
	} else {
		return m.RawCh
	}
}

func (m *SubscribeMessage) Info() string {
	if m.RawClientInfo != "" {
		return m.RawClientInfo
	} else {
		return m.RawCi
	}
}

//
// Publish
//

type PublishMessage struct {
	RawChannel    string `json:"channel"`
	RawCh         string `json:"ch"`
	RawMessage    string `json:"message"`
	RawMsg        string `json:"msg"`
	RawTag        string `json:"tag"`
	RawExtention  string `json:"extention"`
	RawExt        string `json:"ext"`
	RawClientInfo string `json:"client_info"`
	RawCi         string `json:"ci"`
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

func (m *PublishMessage) Info() string {
	if m.RawClientInfo != "" {
		return m.RawClientInfo
	} else {
		return m.RawCi
	}
}

func NewPublishMessage(channel string, ch string, message string, msg string, tag string, extention string, ext string, client_info string, ci string) *PublishMessage {
	pmsg := &PublishMessage{
		RawChannel:    channel,
		RawCh:         ch,
		RawMessage:    message,
		RawMsg:        msg,
		RawTag:        tag,
		RawExtention:  extention,
		RawExt:        ext,
		RawClientInfo: client_info,
		RawCi:         ci,
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
		for i, ri := range rm.GetRoomInfos() {
			remoteAddrs := []string{}
			for _, c := range ri.Room.GetMembers() {
				remoteAddr := c.GetSocket().RemoteAddr().String()
				infoAtRemote := remoteAddr
				if info, exist := cliInfos.Load(remoteAddr); exist {
					if TARGET_PAAS {
						infoAtRemote = info.(string)
					} else {
						infoAtRemote = info.(string) + "@" + remoteAddr
					}
				} else {
					if TARGET_PAAS {
						// mask ip address
						infoAtRemote = fmt.Sprintf("conn_%d", i)
					} else {
						infoAtRemote = remoteAddr
					}
				}

				remoteAddrs = append(remoteAddrs, infoAtRemote)
			}
			channels[ri.Topic] = remoteAddrs
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

//
// Plugin
//

type PluginMessage struct {
	RawCommand string `json:"command"`
	RawCmd     string `json:"cmd"`
}

func (m *PluginMessage) Command() string {
	if m.RawCommand != "" {
		return m.RawCommand
	} else {
		return m.RawCmd
	}
}

func NewPluginMessage(command string, cmd string) *PluginMessage {
	msg := &PluginMessage{
		RawCommand: command,
		RawCmd:     cmd,
	}
	return msg
}
