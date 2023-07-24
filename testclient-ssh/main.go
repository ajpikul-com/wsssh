package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ajpikul-com/wsssh/wsconn"
	gws "github.com/gorilla/websocket"
)

var url string = "ws://127.0.0.1:4648"

func ReadTexts(conn *wsconn.WSConn) {
	textChan := make(chan int)
	conn.TextChan = textChan
	p := make([]byte, 1024)
	for _ = range textChan {
		for conn.TextBuffer.Len() > 0 { // Length left to read
			n, err := conn.TextBuffer.Read(p)
			defaultLogger.Info("ReadTexts: " + string(p[0:n]))
			if err != nil {
				defaultLogger.Error("ReadTexts: TextBuffer.Read():" + err.Error())
				break
			}
		}
	}
	defaultLogger.Info("ReadTexts Channel Closed")
	// The channel has been closed by someone else
}

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

func Pinger(conn *wsconn.WSConn) {
	for {
		err := conn.WritePing([]byte("Perring"))
		if err != nil {
			defaultLogger.Error("Pinger: " + err.Error())
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	// Doing a lot of stuff manually w/ websockets - still not sure why I'm doing it this way
	defaultLogger.Info("Initializing websockets dialer from client")

	_, cancel := context.WithCancel(context.Background())
	defer cancel() // Is this really necessary?
	dialer := gws.Dialer{}
	conn1, resp, err := dialer.Dial(url, nil)
	if err != nil {
		defaultLogger.Error("websocket.Dialier.Dial: Dial fail: " + err.Error())
		dumpResponse(resp)
		return
	}
	wssshConn, err := wsconn.New(conn1)
	if err != nil {
		panic(err.Error())
	}

	var wg sync.WaitGroup
	go ReadTexts(wssshConn)

	// Move these all to flags TODO
	_, err = GetClient(wssshConn, "ajp", "/home/ajp/.ssh/id_ed25519", "/home/ajp/systems/public_keys/ajpikul.com_hostkey")
	if err == nil {
		// Start goroutine to wait for signal to close
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		wg.Add(1)
		go func() {
			for _ = range c {
				break
			}
			wg.Done()
		}()
	}
	wg.Wait()
	wssshConn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
	err = wssshConn.CloseAll()
	if err != nil {
		defaultLogger.Info("Tried to close: " + err.Error())
	}
}
