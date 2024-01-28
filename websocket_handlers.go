package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
	if !IpValidation(r.RemoteAddr) {
		log.Println(fmt.Sprintf("> [Warning] remote ip blocked from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "remote ip blocked", logrus.Fields{"method": "connect", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "remote ip blocked")
		j, _ := json.Marshal(msg)
		conn.Emit("message", string(j))

		go func(c *golem.Connection) {
			time.Sleep(time.Millisecond * 1)
			conn.Close()
		}(conn)

		return
	}

	if opts.SecureMode {
		smsg := SecureHandler(r)
		res, err := Authenticate(secret, smsg.Token(), host)
		if !res || err != nil {
			log.Println(fmt.Sprintf("> [Warning] authentication failed from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "authentication failed", logrus.Fields{"method": "connect", "token": smsg.Token(), "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "security error")
			j, _ := json.Marshal(msg)
			conn.Emit("message", string(j))

			go func(c *golem.Connection) {
				time.Sleep(time.Millisecond * 1)
				conn.Close()
			}(conn)

			return
		}
	}

	remote := SplitAddr(r.RemoteAddr)

	if _, exist := conns[r.RemoteAddr]; exist {
		log.Println(fmt.Sprintf("> [Warning] %s is already connecting", remote))
		logger.Log(WARN, "already connecting", logrus.Fields{"method": "connect", "from": remote})

		time.Sleep(time.Millisecond * 1)
		conn.Close()
	} else {
		conns[r.RemoteAddr] = conn

		log.Println(fmt.Sprintf("> [Connected] from %s", remote))
		if logger != nil {
			logger.Log(INFO, "new connection", logrus.Fields{"method": "connect", "from": remote})
		}
	}
}

func Ping(conn *golem.Connection) {
	conn.Emit("message", "pong")
}

func Subscribe(conn *golem.Connection, msg *SubscribeMessage) {
	remote := GetRemoteIPfromConn(conn)
	infoAtRemote := remote
	if msg.Info() != "" {
		infoAtRemote = msg.Info() + "@" + remote
	}

	if msg.Channel() == "" {
		log.Println(fmt.Sprintf("> [Warning] subscribe channel is empty from %s", infoAtRemote))
		if logger != nil {
			logger.Log(WARN, "subscribe channel is empty", logrus.Fields{"method": "subscribe", "channel": msg.Channel(), "from": infoAtRemote})
		}
		return
	}

	if len(safeList) > 0 {
		contain := false
		for _, ch := range safeList {
			if msg.Channel() == ch {
				contain = true
				break
			}
		}
		if !contain {
			log.Println(fmt.Sprintf("> [Warning] whitelist does not contain subscribe channel from %s", infoAtRemote))
			if logger != nil {
				logger.Log(WARN, "whitelist does not contain subscribe channel", logrus.Fields{"method": "subscribe", "channel": msg.Channel(), "from": infoAtRemote})
			}
			return
		}
	}

	log.Println(fmt.Sprintf("> [Subscribe] ch:%s from %s", msg.Channel(), infoAtRemote))
	if logger != nil {
		logger.Log(INFO, "new subscribe", logrus.Fields{"method": "subscribe", "channel": msg.Channel(), "from": infoAtRemote})
	}

	if msg.Info() != "" {
		cliInfos[remote] = msg.Info()
	}

	roomMg.Join(msg.Channel(), conn)
}

func Unsubscribe(conn *golem.Connection, msg *SubscribeMessage) {
	remote := GetRemoteIPfromConn(conn)
	infoAtRemote := remote
	if msg.Info() != "" {
		infoAtRemote = msg.Info() + "@" + remote
	}

	if msg.Channel() == "" {
		log.Println(fmt.Sprintf("> [Warning] unsubscribe channel is empty from %s", infoAtRemote))
		if logger != nil {
			logger.Log(WARN, "unsubscribe channel is empty", logrus.Fields{"method": "unsubscribe", "channel": msg.Channel(), "from": infoAtRemote})
		}
		return
	}

	log.Println(fmt.Sprintf("> [Unsubscribe] ch:%s from %s", msg.Channel(), infoAtRemote))
	if logger != nil {
		logger.Log(INFO, "unsubscribe", logrus.Fields{"method": "unsubscribe", "channel": msg.Channel(), "from": infoAtRemote})
	}

	if _, exist := cliInfos[remote]; exist {
		delete(cliInfos, remote)
	}

	roomMg.Leave(msg.Channel(), conn)
}

func Publish(conn *golem.Connection, msg *PublishMessage) {
	remote := GetRemoteIPfromConn(conn)
	infoAtRemote := remote
	if info, exist := cliInfos[remote]; exist {
		infoAtRemote = info + "@" + remote
	} else if msg.Info() != "" {
		infoAtRemote = msg.Info() + "@" + remote
	}

	if msg.Channel() == "" {
		log.Println(fmt.Sprintf("> [Warning] publish channel is empty from %s", infoAtRemote))
		if logger != nil {
			logger.Log(WARN, "publish channel is empty", logrus.Fields{"method": "publish", "channel": msg.Channel(), "message": msg.Message(), "tag": msg.Tag(), "extention": msg.Extention(), "from": infoAtRemote})
		}
		return
	}

	log.Println(fmt.Sprintf("> [Publish] ch:%s msg:%s from %s", msg.Channel(), msg.BuildLogString(), infoAtRemote))
	if logger != nil {
		logger.Log(INFO, "new publish", logrus.Fields{"method": "publish", "channel": msg.Channel(), "message": msg.Message(), "tag": msg.Tag(), "extention": msg.Extention(), "from": infoAtRemote})
	}

	pmsg := NewPublishSendMessage(msg.Channel(), msg.Message(), msg.Tag(), msg.Extention())
	if strings.HasSuffix(msg.Channel(), "/*") {
		groupCh := strings.TrimSuffix(msg.Channel(), "/*")
		for _, ri := range roomMg.GetRoomInfos() {
			if strings.HasPrefix(ri.Topic, groupCh) {
				ck := strings.TrimPrefix(ri.Topic, groupCh)
				if len(ck) > 0 && strings.HasPrefix(ck, "/") {
					roomMg.Emit(ri.Topic, "message", &pmsg)
				}
			}
		}
	} else {
		roomMg.Emit(msg.Channel(), "message", &pmsg)
	}
}

func Status(conn *golem.Connection) {
	msg := NewStatusMessage(roomMg)

	conn.Emit("message", &msg)
}

func Closed(conn *golem.Connection) {
	remote := GetRemoteIPfromConn(conn)
	if remote == "" {
		remote = SplitAddr(conn.GetSocket().RemoteAddr().String())
	}
	infoAtRemote := remote
	if info, exist := cliInfos[remote]; exist {
		infoAtRemote = info + "@" + remote
	}

	log.Println(fmt.Sprintf("> [Closed] from %s", infoAtRemote))
	if logger != nil {
		logger.Log(INFO, "connection close", logrus.Fields{"method": "close", "from": infoAtRemote})
	}

	roomMg.LeaveAll(conn)
}
