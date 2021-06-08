package httpprox

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	SourcePort int

	TargetHost     string
	TargetPort     int
	TargetLogin    string
	TargetPassword string
}

var config Config

func handleClient(w http.ResponseWriter, r *http.Request) {
	// log.Printf("r: %v", r)
	logrus.Info(r)
	if r.Method == http.MethodConnect {
		handleConnect(w, r)
	} else {
		handleHttp(w, r)
	}
}

func handleConnect(w http.ResponseWriter, r *http.Request) {

	dest_conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", config.TargetHost, config.TargetPort), 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	s := fmt.Sprintf("%v:%v", config.TargetLogin, config.TargetPassword)
	s = base64.StdEncoding.EncodeToString([]byte(s))
	r.Header.Add("Proxy-Authorization", "Basic "+s)
	r.Write(dest_conn)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func handleHttp(w http.ResponseWriter, req *http.Request) {
	u, err := url.Parse(fmt.Sprintf("http://%v:%v@%v:%v", config.TargetLogin, config.TargetPassword, config.TargetHost, config.TargetPort))
	if err != nil {
		panic(err)
	}
	tr := &http.Transport{Proxy: http.ProxyURL(u)}
	client := &http.Client{Transport: tr}

	resp, err := client.Transport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func Start(superprox string, proxyPort int, hostport int, username string, pwrd string) (error, string) {
	config.TargetLogin = username
	config.TargetPassword = pwrd
	config.TargetHost = superprox
	config.TargetPort = proxyPort
	config.SourcePort = hostport
	server := &http.Server{Addr: fmt.Sprintf(":%v", config.SourcePort), Handler: http.HandlerFunc(handleClient)}
	server.ListenAndServe()
	return nil, "localhost:" + strconv.Itoa(hostport)
}
