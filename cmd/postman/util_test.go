package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/sharkattack51/golem"
	"github.com/stretchr/testify/require"
)

func TestGetRemoteIPfromConn(t *testing.T) {
	opts = Options{}
	Prepare()

	// start server
	s := StartMockServer(t)

	// connect and subscribe
	c := RequireConnectAndSubscribe(t, s.URL, "TEST_CH", "")

	conn, _ := conns.Load(c.LocalAddr().String())
	ip := GetRemoteIPfromConn(conn.(*golem.Connection))

	require.Equal(t, ip, "127.0.0.1")

	ip = GetRemoteIPfromConn(nil)

	require.Equal(t, ip, "")
}

func TestGetHostIP(t *testing.T) {
	// runtime.GOOS() mock for test
	goos := RuntimeGOOS
	defer func() { RuntimeGOOS = goos }()
	RuntimeGOOS = "windows"

	// extract ip4 address
	cmd := exec.Command("sh", "-c", `ifconfig | grep inet | grep -v 127.0.0.1 | cut -d: -f2 | awk '$2 !=""{print $2}'`)
	var buf bytes.Buffer
	stdout := cmd.Stdout
	defer func() { cmd.Stdout = stdout }()
	cmd.Stdout = &buf
	cmd.Run()

	host := GetHostIP()

	found := false
	for _, v := range strings.Split(buf.String(), "\n") {
		if strings.Contains(host, v) {
			found = true
			break
		}
	}
	require.True(t, found || strings.Contains(host, "127.0.0.1"))
}
