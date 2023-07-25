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
	defaultLogger.Debug("Beginning to read texts")
	for s := range channel {
		defaultLogger.Info("ReadTexts: " + s)
	}
	defaultLogger.Debug("ReadTexts Channel Closed")

}

func WriteText(conn *wsconn.WSConn) {
	for {
		_, err := conn.WriteText([]byte("Test Message")) // TODO Can we be sure this will write everything
		if err != nil {
			defaultLogger.Error("wsconn.WriteText(): " + err.Error())
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func Pinger(conn *wsconn.WSConn) error {
	defaultLogger.Debug("Beggining Ping Loop")
	for {
		err := conn.WritePing([]byte("Pingaring'll Payload"))
		if err != nil {
			defaultLogger.Error("Pinger dead: " + err.Error())
			return err
		}
		time.Sleep(10000 * time.Millisecond)
	}
	defaultLogger.Debug("Ending Ping Loop, will never get here")
	return nil
}

func main() {

	var wg sync.WaitGroup
	var wsconn *wsconn.WSConn

	// Start goroutine to wait for signal to close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	wg.Add(1)
	go func() {
		defaultLogger.Debug("Waiting for SIGINT")
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
			defaultLogger.Debug("Trying to reconnect")
			wsconn, err = Reconnect()
			if err != nil {
				defaultLogger.Error("Problem with reconnect: " + err.Error())
				break
			}
			// There is a conn.Wait() but since we're blocking Close() I'm not sure it'll work.
			// I don't know what causes it to get triggered, since Close() is different from server/client ssh.Conn (or maybe they share code/interface) and they don't seem to do more besides calling underlying close, which is blocked. Also, right now, they are the ones who are driving the binary Read() function, which also processes the text Read() function.
			// However, a read error in wsconn could trigger wsconn close which might end up calling wait. Right now we use Pinger to see if things are closed but it seems to buffer. and takes a lwhile.
			go ReadTexts(wsconn)
			err = Pinger(wsconn)
			defaultLogger.Debug("Pinger Error: " + err.Error())
			// Why bother, we can't do this
			wsconn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
			// We're closing all to signal the end of some go routines
			wsconn.CloseAll()
		}
		defaultLogger.Debug("Trying to send myself interrupt")

		pid := os.Getpid()
		p, _ := os.FindProcess(pid)
		_ = p.Signal(os.Interrupt)

	}()

	wg.Wait()
	defaultLogger.Debug("passed wg.Wait()")
	if wsconn != nil {
		defaultLogger.Debug("Trying to close cleanly")
		wsconn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
		err := wsconn.CloseAll()
		if err != nil {
			defaultLogger.Info("Tried to close: " + err.Error())
		}
	}

}
