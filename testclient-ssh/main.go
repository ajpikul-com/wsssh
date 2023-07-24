package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ajpikul-com/wsssh/wsconn"
	gws "github.com/gorilla/websocket"
)

var url string = "ws://127.0.0.1:4648"

func ReadTexts(conn *wsconn.WSConn) {
	channel, _ := conn.SubscribeToTexts()
	defaultLogger.Info("Beginning to read texts")
	for s := range channel {
		defaultLogger.Info("ReadTexts: " + s)
	}
	defaultLogger.Info("ReadTexts Channel Closed")

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

func Pinger(conn *wsconn.WSConn) error {
	defaultLogger.Info("Beggining Ping Loop")
	for {
		err := conn.WritePing([]byte("Pingaring'll Payload"))
		if err != nil {
			defaultLogger.Error("Pinger dead: " + err.Error())
			return err
		}
		time.Sleep(10000 * time.Millisecond)
	}
	defaultLogger.Info("Ending Ping Loop, will never get here")
	return nil
}

func main() {
	// Doing a lot of stuff manually w/ websockets - still not sure why I'm doing it this way
	defaultLogger.Info("Initializing websockets dialer from client")

	var wg sync.WaitGroup
	var conn *wsconn.WSConn

	// Start goroutine to wait for signal to close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	wg.Add(1)
	go func() {
		defaultLogger.Info("Waiting for SIGINT")
		for _ = range c {
			// Here we exit the program
			defaultLogger.Info("Recieved SIGINT")
			break
		}
		wg.Done()
	}()
	go func() {
		for { // All this depends on ssh sitting on top of Read() which TODO Not sure it does
			var err error
			conn, err = Reconnect()
			if err != nil {
				defaultLogger.Error("Problem with reconnect: ")
				defaultLogger.Error(err.Error())
				break
			}
			// There is a conn.Wait() but since we're blocking Close() I'm not sure it'll work
			// But if we get a write error on the underlying ssh, we should be good to go
			go ReadTexts(conn)
			err = Pinger(conn)
			defaultLogger.Error("Pinger Error: ")
			defaultLogger.Error(err.Error())
			conn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
			conn.CloseAll() // close it all, we don't want a memory leak
		}
		defaultLogger.Info("Trying to send myself interrupt")

		pid := os.Getpid()
		p, _ := os.FindProcess(pid)
		_ = p.Signal(os.Interrupt)

	}()

	wg.Wait()
	defaultLogger.Info("passed wg.Wait()")
	if conn != nil {
		defaultLogger.Info("Trying to close cleanly")
		conn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
		err := conn.CloseAll()
		if err != nil {
			defaultLogger.Info("Tried to close: " + err.Error())
		}
	}
	// The channel has been closed by someone else

}
