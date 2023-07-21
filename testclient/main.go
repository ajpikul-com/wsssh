package main

import (
	"context"
	//"strconv"
	//"sync"
	"time"

	"github.com/ajpikul-com/wsssh/wsconn"
	gws "github.com/gorilla/websocket"
)

var url string = "ws://127.0.0.1:2223"

func WriteText(conn *wsconn.WSConn) {
	for {
		_, err := conn.WriteText([]byte("Test Message")) // TODO Can we be sure this will write everything
		if err != nil {
			defaultLogger.Error("WriteText: wsconn.WriteText(): " + err.Error())
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func ReadBinary(conn *wsconn.WSConn) {
	var n int = 0
	readBuffer := make([]byte, 1024)
	var err error = nil
	for err == nil {
		for n, err = conn.Read(readBuffer); n != 0; n, err = conn.Read(readBuffer) {
			defaultLogger.Info("ReadBinary: " + string(readBuffer[0:n]))
			if err != nil {
				// Errors usually won't get here because n = 0
				defaultLogger.Error("ReadBinary: wsconn.Read():" + err.Error())
				break
			}
		}
		// The error we got was fatal
		// It was probably a close close error, and that's fine
		// Either way, everythings fucked and we wait for reconnect
	}
	defaultLogger.Error("ReadBinary closing with err: " + err.Error())
}

func WriteBinary(conn *wsconn.WSConn) {
	for i := 0; i < 3; i++ {
		_, err := conn.Write([]byte("12345678")) // TODO sure it wil write everyting?
		if err != nil {
			defaultLogger.Error("WriteBinary: wsconn.Write(): " + err.Error())
			break
		}
		time.Sleep(2000 * time.Millisecond)
	}
}

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

	go WriteText(wssshConn)
	go ReadBinary(wssshConn)
	WriteBinary(wssshConn)

	wssshConn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
	err = wssshConn.Close()
	if err != nil {
		defaultLogger.Info("Tried to close: " + err.Error())
	}
	// Not we have a wrapped connection, we need to be able to call things on it.
}
