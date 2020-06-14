package main

import (
	"context"
	"log"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mangodx/AccessTunnel/sshoverws"
)

func dialError(url string, resp *http.Response, err error) {
	if resp != nil {
		extra := ""
		if true {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read HTTP body: %v", err)
			}
			extra = "Body:\n" + string(b)
		}
		log.Printf("%s: HTTP error: %d %s\n%s", err, resp.StatusCode, resp.Status, extra)

	}
	log.Printf("Dial to %q fail: %v", url, err)
}

func main() {
	log.Printf("Here")
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	dialer := websocket.Dialer{}
	log.Printf("Made dialer")
	conn, resp, err := dialer.Dial("ws://houston.osoximeter.com",nil)
	if err != nil {
		log.Printf("Err (%s)", err)
	}
	dialError("houston.osoximeter.com", resp, err)
	ioConn := sshoverws.WrapConn(conn)
	defer ioConn.Close()
}
