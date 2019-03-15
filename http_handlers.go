package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

//
// Http Handlers
//

func PublishHandler(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]string)
	query := r.URL.Query()
	for _, s := range []string{"channel", "ch", "message", "msg", "tag", "extention", "ext"} {
		param := query[s]
		if len(param) > 0 {
			params[s] = param[0]
		} else {
			params[s] = ""
		}
	}

	hasQuery := false
	if params["channel"] != "" || params["ch"] != "" {
		hasQuery = true
	}

	// for GET url-param
	msg := NewPublishMessage(params["channel"], params["ch"], params["message"], params["msg"], params["tag"], params["extention"], params["ext"])

	// for POST form-data
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

	if msg.Channel() == "" {
		log.Println(fmt.Sprintf("> [Worning] publish channel is empty from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "publish channel is empty", logrus.Fields{"method": "publish", "channel": msg.Channel(), "message": msg.Message(), "tag": msg.Tag(), "extention": msg.Extention(), "from": r.RemoteAddr})
		}

		res := NewResultMessage("fail", "publish channel is empty")
		j, _ := json.Marshal(res)
		fmt.Fprint(w, string(j))
	} else {
		log.Println(fmt.Sprintf("> [Publish] ch:%s msg:%s from %s", msg.Channel(), msg.BuildLogString(), r.RemoteAddr))
		if logger != nil {
			logger.Log(INFO, "new publish", logrus.Fields{"method": "publish", "channel": msg.Channel(), "message": msg.Message(), "tag": msg.Tag(), "extention": msg.Extention(), "from": r.RemoteAddr})
		}

		pmsg := NewPublishSendMessage(msg.Channel(), msg.Message(), msg.Tag(), msg.Extention())
		if roomMg != nil {
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

		res := NewResultMessage("success", "")
		j, _ := json.Marshal(res)
		fmt.Fprint(w, string(j))
	}
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	msg := NewStatusMessage(roomMg)
	j, _ := json.Marshal(msg)
	fmt.Fprint(w, string(j))
}

func StatusPpHandler(w http.ResponseWriter, r *http.Request) {
	msg := NewStatusMessage(roomMg)
	j, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Fprint(w, string(j))
}

func StoreHandler(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]string)
	query := r.URL.Query()
	for _, s := range []string{"command", "cmd", "key", "value", "val"} {
		param := query[s]
		if len(param) > 0 {
			params[s] = param[0]
		} else {
			params[s] = ""
		}
	}

	hasQuery := false
	if (params["command"] != "" || params["cmd"] != "") && params["key"] != "" {
		hasQuery = true
	}

	// for GET url-param
	msg := NewStoreMessage(params["command"], params["cmd"], params["key"], params["value"], params["val"])

	// for POST form-data
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

	if msg.Command() != "" {
		if msg.Key() != "" {
			switch strings.ToLower(msg.Command()) {
			case "get":
				log.Println(fmt.Sprintf("> [Store] cmd:%s key:%s from %s", msg.Command(), msg.Key(), r.RemoteAddr))
				if logger != nil {
					logger.Log(INFO, "request store get", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "from": r.RemoteAddr})
				}

				v, err := StoreGet(kvsDB, msg)
				if err != nil {
					res := NewResultMessage("", "") // Key無しの場合
					j, _ := json.Marshal(res)
					fmt.Fprint(w, string(j))
					return
				}

				res := NewResultMessage(v, "")
				j, _ := json.Marshal(res)
				fmt.Fprint(w, string(j))

			case "set":
				log.Println(fmt.Sprintf("> [Store] cmd:%s key:%s val:%s from %s", msg.Command(), msg.Key(), msg.Value(), r.RemoteAddr))
				if logger != nil {
					logger.Log(INFO, "request store set", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "val": msg.Value(), "from": r.RemoteAddr})
				}

				err := StoreSet(kvsDB, msg)
				if err != nil {
					res := NewResultMessage("fail", err.Error())
					j, _ := json.Marshal(res)
					fmt.Fprint(w, string(j))
					return
				}

				res := NewResultMessage("success", "")
				j, _ := json.Marshal(res)
				fmt.Fprint(w, string(j))

			case "has":
				log.Println(fmt.Sprintf("> [Store] cmd:%s key:%s from %s", msg.Command(), msg.Key(), r.RemoteAddr))
				if logger != nil {
					logger.Log(INFO, "request store haskey", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "from": r.RemoteAddr})
				}

				b, err := StoreHas(kvsDB, msg)
				if err != nil {
					res := NewResultMessage("fail", err.Error())
					j, _ := json.Marshal(res)
					fmt.Fprint(w, string(j))
					return
				}

				res := NewResultMessage(strconv.FormatBool(b), "")
				j, _ := json.Marshal(res)
				fmt.Fprint(w, string(j))

			case "del":
				log.Println(fmt.Sprintf("> [Store] cmd:%s key:%s from %s", msg.Command(), msg.Key(), r.RemoteAddr))
				if logger != nil {
					logger.Log(INFO, "request store delete", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "from": r.RemoteAddr})
				}

				err := StoreDelete(kvsDB, msg)
				if err != nil {
					res := NewResultMessage("fail", err.Error())
					j, _ := json.Marshal(res)
					fmt.Fprint(w, string(j))
					return
				}

				res := NewResultMessage("success", "")
				j, _ := json.Marshal(res)
				fmt.Fprint(w, string(j))

			default:
				log.Println(fmt.Sprintf("> [Worning] store command nou found from %s", r.RemoteAddr))
				if logger != nil {
					logger.Log(WARN, "command not found", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "value": msg.Value(), "from": r.RemoteAddr})
				}

				res := NewResultMessage("fail", "store command not found")
				j, _ := json.Marshal(res)
				fmt.Fprint(w, string(j))
			}
		} else {
			log.Println(fmt.Sprintf("> [Worning] store key is empty from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "store key is empty", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "value": msg.Value(), "from": r.RemoteAddr})
			}

			res := NewResultMessage("fail", "store key is empty")
			j, _ := json.Marshal(res)
			fmt.Fprint(w, string(j))
		}
	} else {
		log.Println(fmt.Sprintf("> [Worning] store command is empty from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "store command is empty", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "value": msg.Value(), "from": r.RemoteAddr})
		}

		res := NewResultMessage("fail", "store command is empty")
		j, _ := json.Marshal(res)
		fmt.Fprint(w, string(j))
	}
}
