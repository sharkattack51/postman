package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sharkattack51/golem"
	"github.com/sirupsen/logrus"
)

//
// Websocket Handlers
//

func CreateRouter() *golem.Router {
	router := golem.NewRouter()
	router.OnConnect(Connected)
	router.On("ping", Ping)
	router.On("subscribe", Subscribe)
	router.On("unsubscribe", Unsubscribe)
	router.On("publish", Publish)
	router.On("status", Status)
	router.OnClose(Closed)

	return router
}

func Connected(conn *golem.Connection, r *http.Request) {
	remote := splitAddr(r.RemoteAddr)

	_, exist := conns[r.RemoteAddr]
	if exist {
		log.Println(fmt.Sprintf("> [Worning] %s is already connecting", remote))
		logger.Log(WARN, "already connecting", logrus.Fields{"method": "connect", "from": remote})
	}

	conns[r.RemoteAddr] = conn

	log.Println(fmt.Sprintf("> [Connected] from %s", remote))
	if logger != nil {
		logger.Log(INFO, "new connection", logrus.Fields{"method": "connect", "from": remote})
	}
}

func Ping(conn *golem.Connection) {
	conn.Emit("message", "pong")
}

func Subscribe(conn *golem.Connection, msg *SubscribeMessage) {
	remote := getRemoteIPfromConn(conn)

	if msg.Channel == "" {
		log.Println(fmt.Sprintf("> [Worning] subscribe channnel is empty from %s", remote))
		if logger != nil {
			logger.Log(WARN, "subscribe channnel is empty", logrus.Fields{"method": "subscribe", "channel": msg.Channel, "from": remote})
		}
		return
	}

	log.Println(fmt.Sprintf("> [Subscribe] ch:%s from %s", msg.Channel, remote))
	if logger != nil {
		logger.Log(INFO, "new subscribe", logrus.Fields{"method": "subscribe", "channel": msg.Channel, "from": remote})
	}

	roomMg.Join(msg.Channel, conn)
}

func Unsubscribe(conn *golem.Connection, msg *SubscribeMessage) {
	remote := getRemoteIPfromConn(conn)

	if msg.Channel == "" {
		log.Println(fmt.Sprintf("> [Worning] unsubscribe channnel is empty from %s", remote))
		if logger != nil {
			logger.Log(WARN, "unsubscribe channnel is empty", logrus.Fields{"method": "unsubscribe", "channel": msg.Channel, "from": remote})
		}
		return
	}

	log.Println(fmt.Sprintf("> [Unsubscribe] ch:%s from %s", msg.Channel, remote))
	if logger != nil {
		logger.Log(INFO, "unsubscribe", logrus.Fields{"method": "unsubscribe", "channel": msg.Channel, "from": remote})
	}

	roomMg.Leave(msg.Channel, conn)
}

func Publish(conn *golem.Connection, msg *PublishMessage) {
	remote := getRemoteIPfromConn(conn)

	if msg.Channel == "" {
		log.Println(fmt.Sprintf("> [Worning] publish channnel is empty from %s", remote))
		if logger != nil {
			logger.Log(WARN, "publish channnel is empty", logrus.Fields{"method": "publish", "channel": msg.Channel, "message": msg.Message, "tag": msg.Tag, "extention": msg.Extention, "from": remote})
		}
		return
	}

	log.Println(fmt.Sprintf("> [Publish] ch:%s / msg:%s from %s", msg.Channel, msg.BuildLogString(), remote))
	if logger != nil {
		logger.Log(INFO, "new publish", logrus.Fields{"method": "publish", "channel": msg.Channel, "message": msg.Message, "tag": msg.Tag, "extention": msg.Extention, "from": remote})
	}

	if strings.HasSuffix(msg.Channel, "/*") {
		groupCh := strings.TrimSuffix(msg.Channel, "/*")
		for _, ri := range roomMg.GetRoomInfos() {
			if strings.HasPrefix(ri.Topic, groupCh) {
				ck := strings.TrimPrefix(ri.Topic, groupCh)
				if len(ck) > 0 && strings.HasPrefix(ck, "/") {
					roomMg.Emit(ri.Topic, "message", &msg)
				}
			}
		}
	} else {
		roomMg.Emit(msg.Channel, "message", &msg)
	}
}

func Status(conn *golem.Connection) {
	msg := NewStatusMessage(roomMg)

	conn.Emit("message", &msg)
}

func Closed(conn *golem.Connection) {
	remote := getRemoteIPfromConn(conn)

	log.Println(fmt.Sprintf("> [Closed] from %s", remote))
	if logger != nil {
		logger.Log(INFO, "connection close", logrus.Fields{"method": "close", "from": remote})
	}

	roomMg.LeaveAll(conn)
}
