package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func RequireGolemClientProtocolMessageString(t *testing.T, b []byte) string {
	t.Helper()

	require.Contains(t, string(b), "message ")

	j := strings.Split(string(b), "message ")[1]
	j = strings.ReplaceAll(j, "\\", "")

	// as json
	j = strings.TrimLeft(j, `"`)
	j = strings.TrimRight(j, `"`)

	return j
}

func RequireGolemClientProtocolMessage(t *testing.T, b []byte, actual string) {
	t.Helper()

	require.Equal(t, string(b[:8]), "message ")
	var msg PublishMessage
	err := json.Unmarshal(b[8:], &msg)

	require.NoError(t, err)
	require.Equal(t, msg.Message(), actual)
}

func RequireResponseIsSuccess(t *testing.T, b []byte) ResultMessage {
	t.Helper()

	var msg ResultMessage
	err := json.Unmarshal(b, &msg)

	require.NoError(t, err)
	require.Equal(t, msg.Result, "success")

	return msg
}

func RequireResponseIsFail(t *testing.T, b []byte, reason string) ResultMessage {
	t.Helper()

	var msg ResultMessage
	err := json.Unmarshal(b, &msg)

	require.NoError(t, err)
	require.Equal(t, msg.Result, "fail")
	require.Equal(t, msg.Error, reason)

	return msg
}

func StartMockServer(t *testing.T) *httptest.Server {
	t.Helper()

	s := httptest.NewServer(http.HandlerFunc(CreateRouter().Handler()))
	t.Cleanup(func() { s.Close() })
	s.URL = strings.Replace(s.URL, "http", "ws", 1)

	return s
}

func RequireConnectAndSubscribe(t *testing.T, url string, ch string, ci string) *websocket.Conn {
	t.Helper()

	// connect
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	t.Cleanup(func() { c.Close() })

	require.NoError(t, err)

	// send subscribe
	j, _ := json.Marshal(&SubscribeMessage{RawChannel: ch, RawClientInfo: ci})
	sub := "subscribe " + string(j)
	err = c.WriteMessage(websocket.TextMessage, []byte(sub))

	require.NoError(t, err)

	return c
}

func RequireConnectAndPublish(t *testing.T, url string, ch string, msg string, ci string) *websocket.Conn {
	t.Helper()

	// connect
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	t.Cleanup(func() { c.Close() })

	require.NoError(t, err)

	// send publish
	j, _ := json.Marshal(&PublishMessage{RawCh: ch, RawMsg: msg, RawCi: ci})
	pub := "publish " + string(j)
	err = c.WriteMessage(websocket.TextMessage, []byte(pub))

	require.NoError(t, err)

	return c
}

func RequirePublish(t *testing.T, c *websocket.Conn, ch string, msg string, tag string, ext string, ci string) {
	t.Helper()

	// send publish
	j, _ := json.Marshal(&PublishMessage{RawCh: ch, RawMsg: msg, RawTag: tag, RawExt: ext, RawCi: ci})
	pub := "publish " + string(j)
	err := c.WriteMessage(websocket.TextMessage, []byte(pub))

	require.NoError(t, err)
}

func RequireConnectAndPing(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	// connect
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	t.Cleanup(func() { c.Close() })

	require.NoError(t, err)

	// send ping
	ping := "ping {}"
	err = c.WriteMessage(websocket.TextMessage, []byte(ping))

	require.NoError(t, err)

	return c
}

func RequireConnectAndStatus(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	// connect
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	t.Cleanup(func() { c.Close() })

	require.NoError(t, err)

	// send status
	ping := "status {}"
	err = c.WriteMessage(websocket.TextMessage, []byte(ping))

	require.NoError(t, err)

	return c
}

func RequireSecureConnect(t *testing.T, url string, tkn string) *websocket.Conn {
	t.Helper()

	// connect
	c, _, err := websocket.DefaultDialer.Dial(url+"?tkn="+tkn, nil)
	t.Cleanup(func() { c.Close() })

	require.NoError(t, err)

	return c
}

func RequireUnsubscribe(t *testing.T, c *websocket.Conn, ch string, ci string) {
	t.Helper()

	time.Sleep(100 * time.Millisecond) // wait

	// send unsubscribe
	j, _ := json.Marshal(&SubscribeMessage{RawCh: ch, RawCi: ci})
	unsub := "unsubscribe " + string(j)
	err := c.WriteMessage(websocket.TextMessage, []byte(unsub))

	require.NoError(t, err)
}

func ReadFmtPrintOut(t *testing.T, fn func()) string {
	t.Helper()

	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return strings.TrimRight(buf.String(), "\n")
}

func ReadLogPrintOut(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	logWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(logWriter)

	if fn != nil {
		fn()
	}

	return buf.String()
}

func RequireContainsLogFile(t *testing.T, logDir string, expect string, linesAgo int) {
	t.Helper()

	time.Sleep(100 * time.Millisecond) // wait

	b, _ := os.ReadFile(filepath.Join(logDir, LOG_FILE))
	logs := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	require.Contains(t, logs[len(logs)-1-linesAgo], expect)
}

func RequireNotContainsLogFile(t *testing.T, logDir string, expect string, linesAgo int) {
	t.Helper()

	time.Sleep(100 * time.Millisecond) // wait

	b, _ := os.ReadFile(filepath.Join(logDir, LOG_FILE))
	logs := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	require.NotContains(t, logs[len(logs)-1-linesAgo], expect)
}

func RequireContainsStdLog(t *testing.T, fn func(), expect string, linesAgo int) {
	t.Helper()

	var buf bytes.Buffer
	logWriter := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(logWriter)

	if fn != nil {
		fn()
	}

	logs := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Contains(t, logs[len(logs)-1-linesAgo], expect)
}
