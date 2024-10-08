package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

var (
	TEST_CHANNEL = "TEST"
	TEST_MESSAGE = "TEST@MESSAGE"
)

//
// Utilities
//

func GolemClientProtocolString(i interface{}) string {
	j, _ := json.Marshal(&i)
	s := string(j)

	switch i.(type) {
	case *SubscribeMessage:
		s = "subscribe " + s
	case *PublishMessage:
		s = "publish " + s
	}

	return s
}

func RequireGolemClientProtocolMessage(t *testing.T, b []byte, actual string) {
	require.Equal(t, string(b[:8]), "message ")
	var msg PublishMessage
	err := json.Unmarshal(b[8:], &msg)

	require.NoError(t, err)
	require.Equal(t, msg.Message(), actual)
}

func RequireResponseIsSuccess(t *testing.T, b []byte) {
	var msg ResultMessage
	err := json.Unmarshal(b, &msg)

	require.NoError(t, err)
	require.Equal(t, msg.Result, "success")
}

//
// Tests
//

func TestWebSocketAPI(t *testing.T) {
	Prepare()

	// start server
	s := httptest.NewServer(http.HandlerFunc(CreateRouter().Handler()))
	t.Cleanup(func() { s.Close() })
	s.URL = strings.Replace(s.URL, "http", "ws", 1)

	t.Run("basic pub_sub process", func(t *testing.T) {
		// client_1: connect
		c1, _, err := websocket.DefaultDialer.Dial(s.URL, nil)

		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // wait

		// client_1: subscribe
		msg := GolemClientProtocolString(&SubscribeMessage{RawCh: TEST_CHANNEL})
		err = c1.WriteMessage(websocket.TextMessage, []byte(msg))

		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // wait

		// client_2: connect
		c2, _, err := websocket.DefaultDialer.Dial(s.URL, nil)

		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // wait

		// client_2: publish -> cli1[TEST] "TEST@MESSAGE"
		msg = GolemClientProtocolString(&PublishMessage{RawCh: TEST_CHANNEL, RawMsg: TEST_MESSAGE})
		err = c2.WriteMessage(websocket.TextMessage, []byte(msg))

		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // wait

		// client_1: recieve message "TEST@MESSAGE"
		_, rcv, err := c1.ReadMessage()

		require.NoError(t, err)
		RequireGolemClientProtocolMessage(t, rcv, TEST_MESSAGE)

		// client_1: close
		err = c1.Close()

		require.NoError(t, err)

		// client_2: close
		err = c2.Close()

		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // wait
	})

	t.Run("simultaneous connection of multiple clients", func(t *testing.T) {
		num := 100

		for i := 0; i < num; i++ {
			ii := i
			t.Run("client_"+strconv.Itoa(ii), func(t *testing.T) {
				t.Parallel()

				// connect
				c, _, err := websocket.DefaultDialer.Dial(s.URL, nil)

				require.NoError(t, err)

				time.Sleep(100 * time.Millisecond) // wait

				// close
				err = c.Close()

				require.NoError(t, err)
			})
		}

		time.Sleep(100 * time.Millisecond) // wait
	})
}

func TestHttpAPI(t *testing.T) {
	Prepare()

	// start server
	s := httptest.NewServer(http.HandlerFunc(CreateRouter().Handler()))
	t.Cleanup(func() { s.Close() })
	s.URL = strings.Replace(s.URL, "http", "ws", 1)

	// client connect
	c, _, err := websocket.DefaultDialer.Dial(s.URL, nil)

	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // wait

	// client subscribe
	msg := GolemClientProtocolString(&SubscribeMessage{RawCh: TEST_CHANNEL})
	err = c.WriteMessage(websocket.TextMessage, []byte(msg))

	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // wait

	t.Run("get status", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		StatusHandler(w, r)

		require.Equal(t, w.Code, http.StatusOK)

		// response
		var msg StatusMessage
		err := json.Unmarshal(w.Body.Bytes(), &msg)

		require.NoError(t, err)
		require.Equal(t, msg.Version, VERSION)
		require.Contains(t, msg.Channels, TEST_CHANNEL)
		for _, v := range msg.Channels[TEST_CHANNEL] {
			require.Contains(t, v, "127.0.0.1")
		}
	})

	t.Run("get status pp", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		StatusPpHandler(w, r)

		require.Equal(t, w.Code, http.StatusOK)

		// response
		var msg StatusMessage
		err := json.Unmarshal(w.Body.Bytes(), &msg)

		require.NoError(t, err)
		require.Equal(t, msg.Version, VERSION)
		require.Contains(t, msg.Channels, TEST_CHANNEL)
		for _, v := range msg.Channels[TEST_CHANNEL] {
			require.Contains(t, v, "127.0.0.1")
		}
	})

	t.Run("send publish", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		q := r.URL.Query()
		q.Add("ch", TEST_CHANNEL)
		q.Add("msg", TEST_MESSAGE)
		r.URL.RawQuery = q.Encode()
		w := httptest.NewRecorder()
		PublishHandler(w, r)

		// response
		require.Equal(t, w.Code, http.StatusOK)
		RequireResponseIsSuccess(t, w.Body.Bytes())

		time.Sleep(100 * time.Millisecond) // wait

		// client recieve message "TEST@MESSAGE"
		_, rcv, err := c.ReadMessage()

		require.NoError(t, err)
		RequireGolemClientProtocolMessage(t, rcv, TEST_MESSAGE)
	})
}
