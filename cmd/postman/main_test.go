package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	// os.Exit() mock for test
	exit := OsExit
	t.Cleanup(func() { OsExit = exit })
	OsExit = func(code int) {}

	os.Setenv(ENV_SECRET, "SECRET")
	t.Cleanup(func() { os.Unsetenv(ENV_SECRET) })

	opts = Options{GenToken: true}
	s := ReadFmtPrintOut(t, Prepare)

	require.Contains(t, s, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.")
}

func TestGenerateTokenEmpty(t *testing.T) {
	// os.Exit() mock for test
	exit := OsExit
	t.Cleanup(func() { OsExit = exit })

	OsExit = func(code int) {}

	// log.Fatalln() mock for test
	fatal := LogFatalln
	t.Cleanup(func() { LogFatalln = fatal })
	LogFatalln = func(v ...any) { log.Println(v) }

	opts = Options{GenToken: true}
	os.Setenv(ENV_SECRET, "")
	t.Cleanup(func() { os.Unsetenv(ENV_SECRET) })

	RequireContainsStdLog(t, Prepare, "environment variable [SECRET] is empty", 0)
}

func TestPrintInfo(t *testing.T) {
	opts = Options{}
	Prepare()

	s := ReadFmtPrintOut(t, func() {
		PrintInfo()
	})

	out := fmt.Sprintf(`===================================================
[[ Postman v%s ]]
websocket server start... ws://%s:/postman

=== Websocket API ===
[Ping]
<- "ping {}"
[Status]
<- "status {}"
[Subscribe]
<- "subscribe {"ch":"CHANNEL",["ci":"CLIENT_INFO"]}"
[Unsubscribe]
<- "unsubscribe {"ch":"CHANNEL"}"
[Publish]
<- "publish {"ch":"CHANNEL","msg":"MESSAGE",["tag":"TAG","ext":"OTHER"]}"

=== Http API ===
http://%s:/postman
[Status]
(GET) /status
(GET) /status_pp
[Publish]
(GET) /publish?ch=CHANNEL&msg=MESSAGE[&tag=TAG&ext=OTHER&ci=CLIENT_INFO]
(POST) /publish <- json={"ch":"CHANNEL","msg":"MESSAGE",["tag":"TAG","ext":"OTHER","ci":"CLIENT_INFO"]}
===================================================`, VERSION, GetHostIP(), GetHostIP())

	require.Equal(t, s, out)
}

func TestSecurePrintInfo(t *testing.T) {
	opts = Options{SecureMode: true, UseStoreApi: true, UseFileApi: true, UsePluginApi: true}
	Prepare()

	s := ReadFmtPrintOut(t, func() {
		PrintInfo()
	})

	out := fmt.Sprintf(`===================================================
[[ Postman v%s ]]
websocket server start... ws://%s:/postman?tkn=TOKEN

=== Websocket API ===
[Ping]
<- "ping {}"
[Status]
<- "status {}"
[Subscribe]
<- "subscribe {"ch":"CHANNEL",["ci":"CLIENT_INFO"]}"
[Unsubscribe]
<- "unsubscribe {"ch":"CHANNEL"}"
[Publish]
<- "publish {"ch":"CHANNEL","msg":"MESSAGE",["tag":"TAG","ext":"OTHER"]}"

=== Http API ===
http://%s:/postman
[Status]
(GET) /status?tkn=TOKEN
(GET) /status_pp?tkn=TOKEN
[Publish]
(GET) /publish?ch=CHANNEL&msg=MESSAGE[&tag=TAG&ext=OTHER&ci=CLIENT_INFO]&tkn=TOKEN
(POST) /publish <- json={"ch":"CHANNEL","msg":"MESSAGE",["tag":"TAG","ext":"OTHER","ci":"CLIENT_INFO"],"tkn":"TOKEN"}
[Store]
(GET) /store?cmd=(GET|SET|HAS|DEL)&key=KEY[&val=VALUE]&tkn=TOKEN
(POST) /store <- json={"cmd":"(GET|SET|HAS|DEL)","key":"KEY",["val":"VALUE"],"tkn":"TOKEN"}
[File]
(GET) /file/FILE_NAME?tkn=TOKEN
(POST) /file <- file=FILE_BINARY json={"tkn":"TOKEN"}
[Plugin]
(GET) /plugin?cmd=COMMAND&tkn=TOKEN
(POST) /plugin <- json={"cmd":COMMAND,"tkn":"TOKEN"}
===================================================`, VERSION, GetHostIP(), GetHostIP())

	require.Equal(t, s, out)
}

func TestDontStartMultipleInstance(t *testing.T) {
	// os.Exit() mock for test
	exit := OsExit
	t.Cleanup(func() { OsExit = exit })
	OsExit = func(code int) {}

	opts = Options{Port: "8800"}
	l, _ := net.Listen("tcp", ":"+opts.Port)
	t.Cleanup(func() { l.Close() })

	RequireContainsStdLog(t, Prepare, "don't start multiple instance", 0)
}

func TestStartAndShutdownServer(t *testing.T) {
	// os.Exit() mock for test
	exit := OsExit
	t.Cleanup(func() { OsExit = exit })
	OsExit = func(code int) {}

	// log.Fatalln() mock for test
	fatal := LogFatalln
	t.Cleanup(func() { LogFatalln = fatal })
	LogFatalln = func(v ...any) { fmt.Println(v) }

	opts = Options{}
	Prepare()

	require.Nil(t, srv)
	StartServer()
	require.NotNil(t, srv)

	time.Sleep(1000 * time.Millisecond) // wait

	s := ReadFmtPrintOut(t, GracefulShutdown)
	require.Contains(t, s, "Server closed")
}

func TestBrokenDB(t *testing.T) {
	os.Rename(DB_FILE, DB_FILE+"_")
	t.Cleanup(func() {
		os.RemoveAll(DB_FILE)
		os.Rename(DB_FILE+"_", DB_FILE)
	})

	opts = Options{UseStoreApi: true}
	Prepare()

	// break postman.db
	os.Remove(filepath.Join(DB_FILE, "CURRENT"))

	s := ReadLogPrintOut(t, Prepare)

	require.Contains(t, s, `could not open "postman.db"`)
	require.Contains(t, s, `recreated "postman.db"`)
	require.NotNil(t, kvsDB)
}

func TestServeFiles(t *testing.T) {
	os.Rename(SERVE_FILES_DIR, SERVE_FILES_DIR+"_")
	t.Cleanup(func() {
		os.RemoveAll(SERVE_FILES_DIR)
		os.Rename(SERVE_FILES_DIR+"_", SERVE_FILES_DIR)
	})

	opts = Options{UseFileApi: true}
	Prepare()

	require.True(t, IsExist(SERVE_FILES_DIR))
}
