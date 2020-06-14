package main

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
	"github.com/mangodx/AccessTunnel/sshoverws"
)

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("houston.osoximeter.com",nil)
	if err != nil {
		log.Printf("Err (%s)", err)
	}
	defer conn.Close()

	ioConn := sshoverws.WrapConn(conn)
	_ = ioConn
}
