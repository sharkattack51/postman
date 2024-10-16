package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/sharkattack51/golem"
)

//
// Mock for Test
//

var OsExit = os.Exit
var RuntimeGOOS = runtime.GOOS
var LogFatalln = log.Fatalln
var GetRemoteAddr = RemoteAddr

//
// Util Functions
//

func GetHostIP() string {
	ip := "127.0.0.1"
	if RuntimeGOOS == "windows" {
		host, _ := os.Hostname()
		addrs, _ := net.LookupIP(host)
		for _, a := range addrs {
			if ipv4 := a.To4(); ipv4 != nil {
				ip = ipv4.String()
				break
			}
		}
	} else {
		addrs, _ := net.InterfaceAddrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					break
				}
			}
		}
	}

	return ip
}

func GetRemoteIPfromConn(conn *golem.Connection) string {
	ip := ""
	conns.Range(func(a interface{}, c interface{}) bool {
		if c.(*golem.Connection) == conn {
			ip = a.(string)
			return false
		}
		return true
	})

	return SplitAddr(ip)
}

func SplitAddr(ip string) string {
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

func RemoteAddr(r *http.Request) string {
	return r.RemoteAddr
}

func ValidIP4(ip string) bool {
	ip = strings.Trim(ip, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	return re.MatchString(ip)
}

func IpValidation(addr string) bool {
	valid := true
	for _, ip := range ipList {
		if !strings.Contains(addr, ip) {
			valid = false
			break
		}
	}
	return valid
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
