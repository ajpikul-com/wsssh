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
	// Why can't ws be http?
	if err != nil {
		log.Printf("Err (%s)", err)
	}
	dialError("houston.osoximeter.com", resp, err)
	log.Printf("Dialed")
	log.Printf("Wrapping")
	ioConn := sshoverws.WrapConn(conn)
	log.Printf("Wrapped")
	var buffer = make([]byte, 256*64)
	ioConn.Write([]byte("Yes!!!!"))
	ioConn.Read(buffer)
	log.Printf("Buffer: %s", buffer)
	ioConn.Read(buffer)
	log.Printf("Buffer: %s", buffer)
	ioConn.Read(buffer)
	log.Printf("Buffer: %s", buffer)
	ioConn.Read(buffer)
	log.Printf("Buffer: %s", buffer)
	ioConn.Read(buffer)
	log.Printf("Buffer: %s", buffer)
	n, _ := ioConn.Read(buffer)
	log.Printf("Buffer: %s", buffer[:n])
	defer ioConn.Close()
}
