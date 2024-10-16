package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func WebSocketMultipleConnectTester(t *testing.T) {
	opts = Options{}
	Prepare()

	// start server
	s := StartMockServer(t)

	// parallel multiple connection
	for i := 0; i < 100; i++ {
		ii := i
		t.Run("connect cli_"+strconv.Itoa(ii), func(t *testing.T) {
			t.Parallel()

			// connect and subscribe
			RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI_"+strconv.Itoa(ii))
		})
	}

	time.Sleep(1000 * time.Millisecond) // wait for parallel resume
}

func WebSocketPublishTester(t *testing.T) {
	opts = Options{}
	Prepare()

	// start server
	s := StartMockServer(t)

	// client_1: connect and subscribe
	c1 := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI_1")

	// client_2: connect and publish
	RequireConnectAndPublish(t, s.URL, "TEST_CH", "TEST@MESSAGE", "TEST_CLI_2")

	// client_1: recieve message "TEST@MESSAGE"
	c1.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, err := c1.ReadMessage()

	require.NoError(t, err)
	fmt.Println(string(rcv))
	RequireGolemClientProtocolMessage(t, rcv, "TEST@MESSAGE")
}

func WebSocketPublishGroupTester(t *testing.T) {
	opts = Options{}
	Prepare()

	// start server
	s := StartMockServer(t)

	// client_1: connect and subscribe
	c1 := RequireConnectAndSubscribe(t, s.URL, "TEST_CH/1", "TEST_CLI_1")

	// client_2: connect and publish
	c2 := RequireConnectAndSubscribe(t, s.URL, "TEST_CH/2", "TEST_CLI_2")
	RequirePublish(t, c2, "TEST_CH/*", "TEST@MESSAGE", "TEST@TAG", "", "TEST_CLI_2")

	// client_1: recieve message "TEST@MESSAGE"
	c1.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, err := c1.ReadMessage()

	require.NoError(t, err)
	fmt.Println(string(rcv))
	RequireGolemClientProtocolMessage(t, rcv, "TEST@MESSAGE")
}

func WebSocketSubscribeIpValidationTester(t *testing.T) {
	opts = Options{IpAddresses: "192.168.0.1"}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	c := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "")

	c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, err := c.ReadMessage()

	require.NoError(t, err)
	j := RequireGolemClientProtocolMessageString(t, rcv)
	RequireResponseIsFail(t, []byte(j), "remote ip blocked")
}

func WebSocketSubscribeSecureModeFailTester(t *testing.T) {
	opts = Options{SecureMode: true}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	c := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "")

	c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, err := c.ReadMessage()

	require.NoError(t, err)
	j := RequireGolemClientProtocolMessageString(t, rcv)
	RequireResponseIsFail(t, []byte(j), "security error")
}

func WebSocketConnectSecureModeTester(t *testing.T) {
	// os.Exit() mock for test
	exit := OsExit
	t.Cleanup(func() { OsExit = exit })
	OsExit = func(code int) {}

	os.Setenv(ENV_SECRET, "SECRET")
	t.Cleanup(func() { os.Unsetenv(ENV_SECRET) })

	opts = Options{GenToken: true}
	Prepare()

	// generate token
	tkn := ReadFmtPrintOut(t, Prepare)
	tkn = strings.Split(tkn, "genarated token: ")[1]

	opts = Options{SecureMode: true}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect fail
	c := RequireSecureConnect(t, s.URL, "@@@")

	c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, err := c.ReadMessage()

	require.NoError(t, err)
	j := RequireGolemClientProtocolMessageString(t, rcv)
	RequireResponseIsFail(t, []byte(j), "security error")

	c.Close()

	// connect success
	c = RequireSecureConnect(t, s.URL, tkn)

	c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, _ = c.ReadMessage()

	require.NotContains(t, string(rcv), "security error")
}

func WebSocketPingTester(t *testing.T) {
	opts = Options{}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and ping
	c := RequireConnectAndPing(t, s.URL)

	c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, err := c.ReadMessage()

	require.NoError(t, err)
	pong := RequireGolemClientProtocolMessageString(t, rcv)
	require.Contains(t, pong, "pong")
}

func WebSocketSameIpConnectionTester(t *testing.T) {
	t.Run("same ip connection is fail", func(t *testing.T) {
		remoteaddr := GetRemoteAddr
		t.Cleanup(func() { GetRemoteAddr = remoteaddr })
		GetRemoteAddr = func(r *http.Request) string {
			// remoteAddr is ip without port number
			return strings.Split(r.RemoteAddr, ":")[0]
		}

		opts = Options{}
		Prepare()

		// start server
		s := StartMockServer(t)

		// connect and subscribe
		c1 := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI_1")

		// connect and subscribe
		c2 := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI_2")

		c1.SetReadDeadline(time.Now().Add(1 * time.Second))
		i, _, err := c1.ReadMessage()

		require.Equal(t, i, -1) // closed
		require.ErrorContains(t, err, "websocket: close")

		c2.SetReadDeadline(time.Now().Add(1 * time.Second))
		i, _, err = c2.ReadMessage()

		require.Equal(t, i, -1) // closed
		require.ErrorContains(t, err, "websocket: close")
	})

	time.Sleep(3000 * time.Millisecond) // wait

	t.Run("difference ip connection is success", func(t *testing.T) {
		opts = Options{}
		Prepare()

		// start server
		s := StartMockServer(t)

		// connect and subscribe
		c1 := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI_1")

		// connect and subscribe
		c2 := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI_2")

		c1.SetReadDeadline(time.Now().Add(1 * time.Second))
		i, _, err := c1.ReadMessage()

		require.Equal(t, i, -1) // no data timeout as connect
		require.ErrorContains(t, err, "timeout")

		c2.SetReadDeadline(time.Now().Add(1 * time.Second))
		i, _, err = c2.ReadMessage()

		require.Equal(t, i, -1) // no data timeout as connect
		require.ErrorContains(t, err, "timeout")
	})
}

func WebSocketSubscribeChannelEmptyTester(t *testing.T) {
	opts = Options{LogDir: "./log"}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	RequireConnectAndSubscribe(t, s.URL, "", "")

	RequireContainsLogFile(t, "./log", "subscribe channel is empty", 0)
}

func WebSocketSubscribeSafelistTester(t *testing.T) {
	opts = Options{Channels: "TEST_WHITE_CH", LogDir: "./log"}
	Prepare()

	// start server
	s := StartMockServer(t)

	t.Run("channel not contain whitelist is fail", func(t *testing.T) {
		// connect and subscribe
		RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "")

		RequireContainsLogFile(t, "./log", "whitelist does not contain subscribe channel", 0)
	})

	t.Run("channel contain whitelist is success", func(t *testing.T) {
		// connect and subscribe
		RequireConnectAndSubscribe(t, s.URL, "TEST_WHITE_CH", "")

		RequireNotContainsLogFile(t, "./log", "whitelist does not contain subscribe channel", 0)
	})
}

func WebSocketSatatusTester(t *testing.T) {
	opts = Options{}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI")

	// connect and status
	c := RequireConnectAndStatus(t, s.URL)
	c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, rcv, err := c.ReadMessage()

	require.NoError(t, err)
	j := RequireGolemClientProtocolMessageString(t, rcv)

	var msg StatusMessage
	json.Unmarshal([]byte(j), &msg)

	require.Equal(t, msg.Version, VERSION)
}

func WebSocketUnsubscribeTester(t *testing.T) {
	opts = Options{LogDir: "./log"}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	c := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "")

	RequireUnsubscribe(t, c, "TEST_CH", "")
	RequireContainsLogFile(t, "./log", "unsubscribe", 0)
}

func WebSocketUnsubscribeChannelEmptyTester(t *testing.T) {
	opts = Options{LogDir: "./log"}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	c := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "TEST_CLI")

	RequireUnsubscribe(t, c, "", "TEST_CLI")
	RequireContainsLogFile(t, "./log", "unsubscribe channel is empty", 0)
}

func WebSocketPublishChannelEmptyTester(t *testing.T) {
	opts = Options{LogDir: "./log"}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	c := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "")

	// publish
	RequirePublish(t, c, "", "TEST@MESSAGE", "", "", "")

	RequireContainsLogFile(t, "./log", "publish channel is empty", 0)
}

func TestWebSocketApi(t *testing.T) {
	WebSocketSubscribeChannelEmptyTester(t) // must first
	WebSocketSubscribeSafelistTester(t)
	WebSocketUnsubscribeTester(t)
	WebSocketUnsubscribeChannelEmptyTester(t)
	WebSocketPublishChannelEmptyTester(t)

	time.Sleep(1000 * time.Millisecond) // wait

	WebSocketPublishTester(t)
	WebSocketPublishGroupTester(t)
	WebSocketSubscribeIpValidationTester(t)
	WebSocketSubscribeSecureModeFailTester(t)
	WebSocketPingTester(t)
	WebSocketSameIpConnectionTester(t)
	WebSocketSatatusTester(t)
	WebSocketConnectSecureModeTester(t)

	time.Sleep(1000 * time.Millisecond) // wait

	WebSocketMultipleConnectTester(t) // must last
}
