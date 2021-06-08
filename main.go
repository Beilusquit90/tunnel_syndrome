package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"tunnel_syndrome/bronet"
	"tunnel_syndrome/httpprox"
)

func main() {
	var username string
	var password string
	var proxy string
	var proxyPort int
	var hostPort int
	var proxyType string
	flag.StringVar(&password, "pwrd", "", "proxy pwrd")
	flag.StringVar(&username, "user", "", "proxy username")
	flag.StringVar(&proxyType, "type", "", "proxy type")
	flag.StringVar(&proxy, "proxy", "", "your name")
	flag.IntVar(&proxyPort, "port", 0, "proxy port")
	flag.IntVar(&hostPort, "port", 0, "host port")
	flag.Parse()
	if strings.Contains(proxy, "http://") {
		proxy = strings.Replace(proxy, "http://", "", 1)
	}
	if len(proxy) == 0 {
		fmt.Println("Usage: defaults.go -proxy")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if proxyPort == 0 || hostPort == 0 || len(proxy) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	switch proxyType {
	case "Http/s":
		httpprox.Start(proxy, proxyPort, hostPort, username, password)
	case "Socks5":
		bronet.StartSocks5(proxy, proxyPort, hostPort, username, password)
	default:

	}
	fmt.Printf("Hello %s\n", proxy)
}
