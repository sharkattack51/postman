package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sharkattack51/golem"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	VERSION       = "0.9"
	LOG_FILE      = "postman.log"
	DB_FILE       = "postman.db"
	TARGET_HEROKU = false
)

var (
	roomMg    *golem.RoomManager
	conns     map[string]*golem.Connection
	whiteList []string
	logger    *Logger
	kvsDB     *leveldb.DB
)

//
// Main
//

func main() {
	wsport := flag.String("wsport", "8800", "listen websocket port number")
	logDir := flag.String("log", "", "output log location")
	chlist := flag.String("chlist", "", "whitelist for channels")
	flag.Parse()

	// for heroku build
	if TARGET_HEROKU {
		hrkEnvPort := os.Getenv("PORT")
		wsport = &hrkEnvPort
		hrkEnvChlist := os.Getenv("CHLIST")
		chlist = &hrkEnvChlist
	}

	// whitelist for subscribe channnels
	chlistSplits := strings.Split(*chlist, ",")
	for _, ch := range chlistSplits {
		if ch != "" {
			whiteList = append(whiteList, ch)
		}
	}

	if !TARGET_HEROKU {
		// log
		if *logDir != "" {
			logger = NewLogger(*logDir, LOG_FILE)
		}

		// store db
		var err error
		kvsDB, err = leveldb.OpenFile(DB_FILE, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer kvsDB.Close()
	}

	host := getHostIP()
	roomMg = golem.NewRoomManager()
	conns = make(map[string]*golem.Connection)

	fmt.Println("===================================================")
	fmt.Println(fmt.Sprintf("[[ Postman v%s ]]", VERSION))
	fmt.Println(fmt.Sprintf("websocket server start... ws://%s:%s/postman", host, *wsport))
	fmt.Println("")
	fmt.Println("=== Websocket API ===")
	fmt.Println("[Ping]")
	fmt.Println("<- \"ping {}\"")
	fmt.Println("[Status]")
	fmt.Println("<- \"status {}\"")
	fmt.Println("[Subscribe]")
	fmt.Println("<- \"subscribe {\"channel\": \"CHANNEL\"}\"")
	fmt.Println("[Unsubscribe]")
	fmt.Println("<- \"unsubscribe {\"channel\": \"CHANNEL\"}\"")
	fmt.Println("[Publish]")
	fmt.Println("<- \"publish {\"channel\": \"CHANNEL\", \"message\": \"MESSAGE\", [\"tag\": \"TAG\", \"extention\": \"OTHER\"]}\"")
	if kvsDB != nil {
		fmt.Println("[Store]")
		fmt.Println("<- \"store {\"command\": \"(GET|SET|HAS|DEL)\", \"key\": \"KEY\", [\"value\": \"VALUE\"]}\"")
	}
	fmt.Println("")
	fmt.Println("=== Http API ===")
	fmt.Println(fmt.Sprintf("http://%s:%s/postman", host, *wsport))
	fmt.Println("[Status]")
	fmt.Println("(GET) /status")
	fmt.Println("(GET) /status_pp")
	fmt.Println("[Publish]")
	fmt.Println("(GET) /publish?channel=CHANNEL&message=MESSAGE[&tag=TAG&extention=OTHER]")
	fmt.Println("(POST) /publish <- \"json={\"channel\": \"CHANNEL\", \"message\": \"MESSAGE\", [\"tag\": \"TAG\", \"extention\": \"OTHER\"]}\"")
	if kvsDB != nil {
		fmt.Println("[Store]")
		fmt.Println("(GET) /store?command=(GET|SET|HAS|DEL)&key=KEY[&value=VALUE]")
		fmt.Println("(POST) /store <- \"json={\"command\": \"(GET|SET|HAS|DEL)\", \"key\": \"KEY\", [\"value\": \"VALUE\"]}\"")
	}
	fmt.Println("===================================================")
	fmt.Println("and all parameters have shortcut string...")
	fmt.Println("[Pub/Sub] \"channnel\"=\"ch\", \"message\"=\"msg\", \"extention\"=\"ext\"")
	if kvsDB != nil {
		fmt.Println("[Store] \"command\"=\"cmd\", \"value\"=\"val\"")
	}
	fmt.Println("")

	if logger != nil {
		logger.Log(INFO, "postman start", logrus.Fields{"host": host, "port": wsport})
	}

	srv := &http.Server{Addr: ":" + (*wsport)}

	// websocket routing
	http.HandleFunc("/postman", CreateRouter().Handler())

	// http routing
	http.HandleFunc("/postman/publish", PublishHandler)
	http.HandleFunc("/postman/status", StatusHandler)
	http.HandleFunc("/postman/status_pp", StatusPpHandler)
	http.HandleFunc("/postman/store", StoreHandler)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sigCh)
	<-sigCh // blocking

	// graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
