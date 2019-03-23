package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/sharkattack51/golem"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	VERSION       = "1.0 alpha"
	LOG_FILE      = "postman.log"
	DB_FILE       = "postman.db"
	TARGET_HEROKU = false

	ENV_SECRET = "SECRET"
	ENV_PORT   = "PORT"
	ENV_CHLIST = "CHLIST"
	ENV_IPLIST = "IPLIST"
)

type Options struct {
	Port        string `short:"p" long:"port" default:"8800" description:"listen port number"`
	LogDir      string `short:"l" long:"log" description:"output log location"`
	Channels    string `short:"c" long:"chlist" description:"whitelist for channels"`
	IpAddresses string `short:"i" long:"iplist" description:"connectable ip_address list"`
	SecureMode  bool   `short:"s" long:"secure" description:"secure mode"`
	GenToken    bool   `short:"g" long:"generate" description:"genarate token from environment variable [POSTMAN_SECRET]"`
}

var (
	host      string
	roomMg    *golem.RoomManager
	conns     map[string]*golem.Connection
	whiteList []string
	ipList    []string
	logger    *Logger
	kvsDB     *leveldb.DB
	opts      Options
	secret    string
)

//
// Main
//

func main() {
	host = GetHostIP()
	roomMg = golem.NewRoomManager()
	conns = make(map[string]*golem.Connection)

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	// generate token mode
	secret = os.Getenv(ENV_SECRET)
	if opts.GenToken {
		if secret == "" {
			log.Fatal(errors.New("environment variable [" + ENV_SECRET + "] is empty"))
		}

		token, err := GenerateToken(secret, host)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("genarated token: " + token)
		os.Exit(0)
	}

	// for heroku build
	if TARGET_HEROKU {
		opts.Port = os.Getenv(ENV_PORT)
		opts.Channels = os.Getenv(ENV_CHLIST)
		opts.IpAddresses = os.Getenv(ENV_IPLIST)
	}

	// whitelist for subscribe channnels
	chSplits := strings.Split(opts.Channels, ",")
	for _, ch := range chSplits {
		if ch != "" {
			whiteList = append(whiteList, ch)
		}
	}

	// iplist for secure connection
	ipSplits := strings.Split(opts.IpAddresses, ",")
	for _, ip := range ipSplits {
		if ip != "" && ValidIP4(ip) {
			ipList = append(ipList, ip)
		}
	}

	if !TARGET_HEROKU {
		// log
		if opts.LogDir != "" {
			logger = NewLogger(opts.LogDir, LOG_FILE)
		}

		// store db
		var err error
		kvsDB, err = leveldb.OpenFile(DB_FILE, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer kvsDB.Close()
	}

	fmt.Println("===================================================")
	fmt.Println(fmt.Sprintf("[[ Postman v%s ]]", VERSION))
	fmt.Println(SecureSprintf(fmt.Sprintf("websocket server start... ws://%s:%s/postman", host, opts.Port)+"%s", "?tkn=TOKEN"))
	fmt.Println("")
	fmt.Println("=== Websocket API ===")
	fmt.Println("[Ping]")
	fmt.Println("<- \"ping {}\"")
	fmt.Println("[Status]")
	fmt.Println("<- \"status {}\"")
	fmt.Println("[Subscribe]")
	fmt.Println("<- \"subscribe {\"ch\":\"CHANNEL\"}\"")
	fmt.Println("[Unsubscribe]")
	fmt.Println("<- \"unsubscribe {\"ch\":\"CHANNEL\"}\"")
	fmt.Println("[Publish]")
	fmt.Println("<- \"publish {\"ch\":\"CHANNEL\",\"msg\":\"MESSAGE\"[,\"tag\":\"TAG\",\"ext\":\"OTHER\"]}\"")
	if kvsDB != nil {
		fmt.Println("[Store]")
		fmt.Println("<- \"store {\"cmd\":\"(GET|SET|HAS|DEL)\",\"key\":\"KEY\"[,\"val\":\"VALUE\"]}\"")
	}
	fmt.Println("")
	fmt.Println("=== Http API ===")
	fmt.Println(fmt.Sprintf("http://%s:%s/postman", host, opts.Port))
	fmt.Println("[Status]")
	fmt.Println(SecureSprintf("(GET) /status%s", "?tkn=TOKEN"))
	fmt.Println(SecureSprintf("(GET) /status_pp%s", "?tkn=TOKEN"))
	fmt.Println("[Publish]")
	fmt.Println(SecureSprintf("(GET) /publish?ch=CHANNEL&msg=MESSAGE[&tag=TAG&ext=OTHER]%s", "&tkn=TOKEN"))
	fmt.Println(SecureSprintf("(POST) /publish <- \"json={\"ch\":\"CHANNEL\",\"msg\":\"MESSAGE\"[,\"tag\":\"TAG\",\"ext\":\"OTHER\"]%s}\"", ",\"tkn\":\"TOKEN\""))
	if kvsDB != nil {
		fmt.Println("[Store]")
		fmt.Println(SecureSprintf("(GET) /store?cmd=(GET|SET|HAS|DEL)&key=KEY[&val=VALUE]%s", "&tkn=TOKEN"))
		fmt.Println(SecureSprintf("(POST) /store <- \"json={\"cmd\":\"(GET|SET|HAS|DEL)\",\"key\":\"KEY\"[,\"val\":\"VALUE\"]%s}\"", ",\"tkn\":\"TOKEN\""))
	}
	fmt.Println("===================================================")
	fmt.Println("")

	if logger != nil {
		logger.Log(INFO, "postman start", logrus.Fields{"host": host, "port": opts.Port})
	}

	srv := &http.Server{Addr: ":" + opts.Port}

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

func SecureSprintf(s string, ss string) string {
	if opts.SecureMode {
		return fmt.Sprintf(s, ss)
	} else {
		return fmt.Sprintf(s, "")
	}
}
