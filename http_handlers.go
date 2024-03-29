package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

//
// Http Handlers
//

func PublishHandler(w http.ResponseWriter, r *http.Request) {
	AllowCORS(w)

	if !IpValidation(r.RemoteAddr) {
		log.Println(fmt.Sprintf("> [Warning] remote ip blocked from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "remote ip blocked", logrus.Fields{"method": "connect", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "remote ip blocked")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if opts.SecureMode {
		smsg := SecureHandler(r)
		res, err := Authenticate(secret, smsg.Token(), host)
		if !res || err != nil {
			log.Println(fmt.Sprintf("> [Warning] authentication failed from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "authentication failed", logrus.Fields{"method": "publish", "token": smsg.Token(), "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "security error")
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}
	}

	params := make(map[string]string)
	query := r.URL.Query()
	for _, s := range []string{"channel", "ch", "message", "msg", "tag", "extention", "ext", "client_info", "ci"} {
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
	msg := NewPublishMessage(params["channel"], params["ch"], params["message"], params["msg"], params["tag"], params["extention"], params["ext"], params["client_info"], params["ci"])

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

	remote := r.RemoteAddr
	infoAtRemote := remote
	if msg.Info() != "" {
		infoAtRemote = msg.Info() + "@" + remote
	}

	if msg.Channel() == "" {
		log.Println(fmt.Sprintf("> [Warning] publish channel is empty from %s", infoAtRemote))
		if logger != nil {
			logger.Log(WARN, "publish channel is empty", logrus.Fields{"method": "publish", "channel": msg.Channel(), "message": msg.Message(), "tag": msg.Tag(), "extention": msg.Extention(), "from": infoAtRemote})
		}

		res := NewResultMessage("fail", "publish channel is empty")
		j, _ := json.Marshal(res)
		fmt.Fprint(w, string(j))
	} else {
		log.Println(fmt.Sprintf("> [Publish] ch:%s msg:%s from %s", msg.Channel(), msg.BuildLogString(), infoAtRemote))
		if logger != nil {
			logger.Log(INFO, "new publish", logrus.Fields{"method": "publish", "channel": msg.Channel(), "message": msg.Message(), "tag": msg.Tag(), "extention": msg.Extention(), "from": infoAtRemote})
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
	AllowCORS(w)

	if !IpValidation(r.RemoteAddr) {
		log.Println(fmt.Sprintf("> [Warning] remote ip blocked from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "remote ip blocked", logrus.Fields{"method": "connect", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "remote ip blocked")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if opts.SecureMode {
		smsg := SecureHandler(r)
		res, err := Authenticate(secret, smsg.Token(), host)
		if !res || err != nil {
			log.Println(fmt.Sprintf("> [Warning] authentication failed from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "authentication failed", logrus.Fields{"method": "status", "token": smsg.Token(), "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "security error")
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}
	}

	log.Println(fmt.Sprintf("> [Status] get status from %s", r.RemoteAddr))
	if logger != nil {
		logger.Log(INFO, "get status", logrus.Fields{"method": "status", "from": r.RemoteAddr})
	}

	msg := NewStatusMessage(roomMg)
	j, _ := json.Marshal(msg)
	fmt.Fprint(w, string(j))
}

func StatusPpHandler(w http.ResponseWriter, r *http.Request) {
	AllowCORS(w)

	if !IpValidation(r.RemoteAddr) {
		log.Println(fmt.Sprintf("> [Warning] remote ip blocked from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "remote ip blocked", logrus.Fields{"method": "connect", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "remote ip blocked")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if opts.SecureMode {
		smsg := SecureHandler(r)
		res, err := Authenticate(secret, smsg.Token(), host)
		if !res || err != nil {
			log.Println(fmt.Sprintf("> [Warning] authentication failed from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "authentication failed", logrus.Fields{"method": "status_pp", "token": smsg.Token(), "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "security error")
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}
	}

	log.Println(fmt.Sprintf("> [Status] get status pp from %s", r.RemoteAddr))
	if logger != nil {
		logger.Log(INFO, "get status pp", logrus.Fields{"method": "status_pp", "from": r.RemoteAddr})
	}

	msg := NewStatusMessage(roomMg)
	j, _ := json.MarshalIndent(msg, "", "    ")
	fmt.Fprint(w, string(j))
}

func StoreHandler(w http.ResponseWriter, r *http.Request) {
	AllowCORS(w)

	if !IpValidation(r.RemoteAddr) {
		log.Println(fmt.Sprintf("> [Warning] remote ip blocked from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "remote ip blocked", logrus.Fields{"method": "connect", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "remote ip blocked")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if opts.SecureMode {
		smsg := SecureHandler(r)
		res, err := Authenticate(secret, smsg.Token(), host)
		if !res || err != nil {
			log.Println(fmt.Sprintf("> [Warning] authentication failed from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "authentication failed", logrus.Fields{"method": "store", "token": smsg.Token(), "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "security error")
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}
	}

	if !opts.UseStoreApi || kvsDB == nil {
		log.Println(fmt.Sprintf("> [Warning] key-value store api is disable from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "key-value store api is disable", logrus.Fields{"method": "file", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "key-value store api is disable")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

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
				log.Println(fmt.Sprintf("> [Warning] store command nou found from %s", r.RemoteAddr))
				if logger != nil {
					logger.Log(WARN, "command not found", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "value": msg.Value(), "from": r.RemoteAddr})
				}

				res := NewResultMessage("fail", "store command not found")
				j, _ := json.Marshal(res)
				fmt.Fprint(w, string(j))
			}
		} else {
			log.Println(fmt.Sprintf("> [Warning] store key is empty from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "store key is empty", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "value": msg.Value(), "from": r.RemoteAddr})
			}

			res := NewResultMessage("fail", "store key is empty")
			j, _ := json.Marshal(res)
			fmt.Fprint(w, string(j))
		}
	} else {
		log.Println(fmt.Sprintf("> [Warning] store command is empty from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "store command is empty", logrus.Fields{"method": "store", "command": msg.Command(), "key": msg.Key(), "value": msg.Value(), "from": r.RemoteAddr})
		}

		res := NewResultMessage("fail", "store command is empty")
		j, _ := json.Marshal(res)
		fmt.Fprint(w, string(j))
	}
}

func SecureHandler(r *http.Request) *SecureMessage {
	params := make(map[string]string)
	query := r.URL.Query()
	for _, s := range []string{"token", "tkn", "password", "pwd"} {
		param := query[s]
		if len(param) > 0 {
			params[s] = param[0]
		} else {
			params[s] = ""
		}
	}

	hasQuery := false
	if params["token"] != "" || params["tkn"] != "" {
		hasQuery = true
	}

	// for GET url-param
	msg := NewSecureMessage(params["token"], params["tkn"])

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

	return msg
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	AllowCORS(w)

	if !IpValidation(r.RemoteAddr) {
		log.Println(fmt.Sprintf("> [Warning] remote ip blocked from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "remote ip blocked", logrus.Fields{"method": "connect", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "remote ip blocked")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if opts.SecureMode {
		smsg := SecureHandler(r)
		res, err := Authenticate(secret, smsg.Token(), host)
		if !res || err != nil {
			log.Println(fmt.Sprintf("> [Warning] authentication failed from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "authentication failed", logrus.Fields{"method": "store", "token": smsg.Token(), "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "security error")
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}
	}

	if !opts.UseFileApi {
		log.Println(fmt.Sprintf("> [Warning] file server api is disable from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "file server api is disable", logrus.Fields{"method": "file", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "file server api is disable")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if !IsExist(SERVE_FILES_DIR) {
		log.Println(fmt.Sprintf("> [Warning] directory not found \"%s\" from %s", SERVE_FILES_DIR, r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, fmt.Sprintf("directory not found \"%s\"", SERVE_FILES_DIR), logrus.Fields{"method": "file post", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", fmt.Sprintf("directory not found \"%s\"", SERVE_FILES_DIR))
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if r.Method == "POST" {
		formFile, header, err := r.FormFile("file")
		defer formFile.Close()

		if err != nil {
			log.Println(fmt.Sprintf("> [Warning] no form file data from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "no form file data", logrus.Fields{"method": "file post", "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "no form file data")
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}

		path := filepath.Join(SERVE_FILES_DIR, header.Filename)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(fmt.Sprintf("> [Warning] could not open file \"%s\" from %s", header.Filename, r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "could not open file", logrus.Fields{"method": "file post", "file": header.Filename, "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", fmt.Sprintf("could not open file \"%s\"", header.Filename))
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}
		defer file.Close()

		data := make([]byte, 1024)
		offset := int64(0)
		for {
			n, err := formFile.Read(data)
			if n == 0 {
				break
			}
			if err != nil {
				break
			}

			// write file
			file.WriteAt(data, offset)
			offset += int64(n)
		}

		log.Println(fmt.Sprintf("> [File] new file posted \"%s\" from %s", header.Filename, r.RemoteAddr))
		if logger != nil {
			logger.Log(INFO, fmt.Sprintf("new file posted"), logrus.Fields{"method": "file post", "file": header.Filename, "from": r.RemoteAddr})
		}

		msg := NewResultMessage("success", "")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))

	} else if r.Method == "GET" {
		urls := strings.Split(r.URL.Path, "/postman/file/")
		pathToFile := ""
		if len(urls) >= 2 {
			pathToFile = urls[1]
		}
		pathToFile = html.UnescapeString(pathToFile)
		if pathToFile == "" {
			pathToFile = "index.html"
		}

		path := filepath.Join(SERVE_FILES_DIR, pathToFile)

		if !IsExist(path) {
			log.Println(fmt.Sprintf("> [Warning] file not found \"%s\" from %s", pathToFile, r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "file not found", logrus.Fields{"method": "file get", "name": pathToFile, "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", fmt.Sprintf("file not found \"%s\"", pathToFile))
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}

		log.Println(fmt.Sprintf("> [File] serve access \"%s\" from %s", pathToFile, r.RemoteAddr))
		if logger != nil {
			logger.Log(INFO, fmt.Sprintf("file served"), logrus.Fields{"method": "file get", "file": pathToFile, "from": r.RemoteAddr})
		}

		// return file
		http.ServeFile(w, r, path)
	}
}

func PluginHandler(w http.ResponseWriter, r *http.Request) {
	AllowCORS(w)

	if !IpValidation(r.RemoteAddr) {
		log.Println(fmt.Sprintf("> [Warning] remote ip blocked from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "remote ip blocked", logrus.Fields{"method": "connect", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "remote ip blocked")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	if opts.SecureMode {
		smsg := SecureHandler(r)
		res, err := Authenticate(secret, smsg.Token(), host)
		if !res || err != nil {
			log.Println(fmt.Sprintf("> [Warning] authentication failed from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "authentication failed", logrus.Fields{"method": "store", "token": smsg.Token(), "from": r.RemoteAddr})
			}

			msg := NewResultMessage("fail", "security error")
			j, _ := json.Marshal(msg)
			fmt.Fprint(w, string(j))
			return
		}
	}

	if !opts.UsePluginApi {
		log.Println(fmt.Sprintf("> [Warning] plugin api is disable from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "plugin api is disable", logrus.Fields{"method": "plugin", "from": r.RemoteAddr})
		}

		msg := NewResultMessage("fail", "plugin api is disable")
		j, _ := json.Marshal(msg)
		fmt.Fprint(w, string(j))
		return
	}

	params := make(map[string]string)
	query := r.URL.Query()
	for _, s := range []string{"command", "cmd"} {
		param := query[s]
		if len(param) > 0 {
			params[s] = param[0]
		} else {
			params[s] = ""
		}
	}

	hasQuery := false
	if params["command"] != "" || params["cmd"] != "" {
		hasQuery = true
	}

	// for GET url-param
	msg := NewPluginMessage(params["command"], params["cmd"])

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
		p, err := LoadPlugin()
		if err != nil {
			log.Println(fmt.Sprintf("> [Warning] plugin can't loaded from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "plugin can't loaded", logrus.Fields{"method": "plugin", "command": msg.Command(), "from": r.RemoteAddr})
			}

			res := NewResultMessage("fail", "plugin can't loaded")
			j, _ := json.Marshal(res)
			fmt.Fprint(w, string(j))
		}

		if proc, ok := p.Plugins[msg.Command()]; ok {
			log.Println(fmt.Sprintf("> [Plugin] call plugin \"%s\" from %s", msg.Command(), r.RemoteAddr))
			if logger != nil {
				logger.Log(INFO, fmt.Sprintf("call plugin"), logrus.Fields{"method": "plugin", "command": msg.Command(), "from": r.RemoteAddr})
			}

			ret := ExecPlugin(proc.Proc, proc.Args)
			fmt.Fprint(w, ret)
		} else {
			log.Println(fmt.Sprintf("> [Warning] plugin command not found from %s", r.RemoteAddr))
			if logger != nil {
				logger.Log(WARN, "plugin command not found", logrus.Fields{"method": "plugin", "command": msg.Command(), "from": r.RemoteAddr})
			}

			res := NewResultMessage("fail", "plugin command not found")
			j, _ := json.Marshal(res)
			fmt.Fprint(w, string(j))
		}
	} else {
		log.Println(fmt.Sprintf("> [Warning] plugin command is empty from %s", r.RemoteAddr))
		if logger != nil {
			logger.Log(WARN, "plugin command is empty", logrus.Fields{"method": "plugin", "command": msg.Command(), "from": r.RemoteAddr})
		}

		res := NewResultMessage("fail", "plugin command is empty")
		j, _ := json.Marshal(res)
		fmt.Fprint(w, string(j))
	}
}

func AllowCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}
