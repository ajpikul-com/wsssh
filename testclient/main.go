package main

import (
	"context"
	"time"

	"github.com/ajpikul-com/wsssh/wsconn"
	gws "github.com/gorilla/websocket"
)

var url string = "ws://127.0.0.1:2223"

func main() {
	// Doing a lot of stuff manually w/ websockets - still not sure why I'm doing it this way
	defaultLogger.Info("Initializing websockets dialer from client")

	_, cancel := context.WithCancel(context.Background())
	defer cancel() // Is this really necessary?
	dialer := gws.Dialer{}
	conn, resp, err := dialer.Dial(url, nil)
	if err != nil {
		defaultLogger.Error("websocket.Dialier.Dial: Dial fail: " + err.Error())
		dumpResponse(resp)
		return
	}
	wssshConn, err := wsconn.New(conn)
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < 2; i++ {
		wssshConn.WriteText([]byte("Test Message"))
		time.Sleep(2000 * time.Millisecond)
		_, err = wssshConn.Write([]byte("12345678"))
		time.Sleep(2000 * time.Millisecond)
	}
	wssshConn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
	err = wssshConn.Close()
	if err != nil {
		defaultLogger.Info("Tried to close: " + err.Error())
	}
	// Not we have a wrapped connection, we need to be able to call things on it.
}
