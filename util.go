package main

import (
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/sharkattack51/golem"
)

//
// Util Functions
//

func getHostIP() string {
	ip := "127.0.0.1"
	if runtime.GOOS == "windows" {
		host, _ := os.Hostname()
		addrs, _ := net.LookupIP(host)
		for _, a := range addrs {
			if ipv4 := a.To4(); ipv4 != nil {
				ip = ipv4.String()
			}
		}
	} else {
		addrs, _ := net.InterfaceAddrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
				}
			}
		}
	}

	return ip
}

func getRemoteIPfromConn(conn *golem.Connection) string {
	ip := ""
	for a, c := range conns {
		if c == conn {
			ip = a
			break
		}
	}

	return splitAddr(ip)
}

func splitAddr(ip string) string {
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}
