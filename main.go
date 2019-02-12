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
)

const (
	VERSION       = "0.8.9"
	LOG_FILE      = "postman.log"
	TARGET_HEROKU = false
)

var (
	roomMg    *golem.RoomManager
	conns     map[string]*golem.Connection
	whiteList []string
	logger    *Logger
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

	// log
	if *logDir != "" {
		logger = NewLogger(*logDir, LOG_FILE)
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
	fmt.Println("<- \"publish {\"channel\": \"CHANNEL\", \"message\": \"MESSAGE\"}\"")
	fmt.Println("<- \"publish {\"channel\": \"CHANNEL\", \"message\": \"MESSAGE\", \"tag\": \"TAG\", \"extention\": \"OTHER\"}\"")
	fmt.Println("")
	fmt.Println("=== Http API ===")
	fmt.Println("[Status]")
	fmt.Println(fmt.Sprintf("GET http://%s:%s/postman/status", host, *wsport))
	fmt.Println(fmt.Sprintf("GET http://%s:%s/postman/status_pp", host, *wsport))
	fmt.Println("[Publish]")
	fmt.Println(fmt.Sprintf("GET http://%s:%s/postman/publish?channel=CHANNEL&message=MESSAGE", host, *wsport))
	fmt.Println(fmt.Sprintf("GET http://%s:%s/postman/publish?channel=CHANNEL&message=MESSAGE&tag=TAG&extention=OTHER", host, *wsport))
	fmt.Println(fmt.Sprintf("POST http://%s:%s/postman/publish <- \"json={\"channel\": \"CHANNEL\", \"message\": \"MESSAGE\"}\"", host, *wsport))
	fmt.Println(fmt.Sprintf("POST http://%s:%s/postman/publish <- \"json={\"channel\": \"CHANNEL\", \"message\": \"MESSAGE\", \"tag\": \"TAG\", \"extention\": \"OTHER\"}\"", host, *wsport))
	fmt.Println("===================================================")

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
