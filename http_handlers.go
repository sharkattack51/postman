package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

//
// Http Handlers
//

func PublishHandler(w http.ResponseWriter, r *http.Request) {
	msg := &PublishMessage{}

	q := r.URL.Query()
	hasQuery := false
	if q_ch, ok := q["channel"]; ok {
		if len(q_ch) > 0 {
			if q_ch[0] != "" {
				hasQuery = true

				msg.Channel = q_ch[0]
				q_msg := q["message"]
				if len(q_msg) > 0 {
					msg.Message = q_msg[0]
				}
				q_tag := q["tag"]
				if len(q_tag) > 0 {
					msg.Tag = q_tag[0]
				}
				q_ext := q["extention"]
				if len(q_ext) > 0 {
					msg.Extention = q_ext[0]
				}
			}
		}
	}

	if !hasQuery {
		r.ParseForm()
		if len(r.Form) > 0 {
			if data, ok := r.Form["json"]; ok {
				if len(data) > 0 {
					json.Unmarshal([]byte(data[0]), msg)
				}
			}
		}
	}

	if msg.Channel == "" {
		log.Println(fmt.Sprintf("> [Worning] publish channnel is empty from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "publish channnel is empty", logrus.Fields{"method": "publish", "channel": msg.Channel, "message": msg.Message, "tag": msg.Tag, "extention": msg.Extention, "from": r.RemoteAddr})
		}
	} else {
		log.Println(fmt.Sprintf("> [Publish] ch:%s / msg:%s from %s", msg.Channel, msg.BuildLogString(), r.RemoteAddr))
		if logger != nil {
			logger.Log(INFO, "new publish", logrus.Fields{"method": "publish", "channel": msg.Channel, "message": msg.Message, "tag": msg.Tag, "extention": msg.Extention, "from": r.RemoteAddr})
		}

		if roomMg != nil {
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
	}

	j, _ := json.Marshal(msg)
	fmt.Fprint(w, string(j))
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	msg := NewStatusMessage(roomMg)

	j, _ := json.Marshal(msg)
	fmt.Fprint(w, string(j))
}
