package main

import (
	"context"
	"io/ioutil"
	"strconv"
	"net/http"
	"golang.org/x/crypto/ssh"

	"github.com/ayjayt/ilog"
	"github.com/gorilla/websocket"
	"github.com/ayjayt/AccessTunnel/sshoverws"
)

var defaultLogger ilog.LoggerInterface

func init(){
	defaultLogger = new(ilog.ZapWrap)
	defaultLogger.(*ilog.ZapWrap).Paths = []string{"./client.log"}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	sshoverws.SetDefaultLogger(defaultLogger)
}

// dialError dumps the boddy in case of an error
func dialError(url string, resp *http.Response, err error) {
	if resp != nil {
		extra := ""
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				defaultLogger.Error("dialError/ioutil.ReadAll: Failed to read HTTP body: "+ err.Error())
				// return ?
			}
			extra = "Body:\n" + string(b)
			defaultLogger.Error("HTTP Response Info: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status + "\n" + extra)

	}
}

func main() {
	config := &ssh.ServerConfig{
		NoClientAuth: true, // TODO lol what but wait
	}

	privateBytes, err := ioutil.ReadFile("/home/ajp/.ssh/id_ed25519")
	if err != nil {
		defaultLogger.Error("ioutil.ReadFile: You must set a proper hostkey with -hostkey: " + err.Error())
		return
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		defaultLogger.Error("ssh.ParsePrivateKey: Failed to parse private key: " + err.Error())
		return
	}

	config.AddHostKey(private)


	// Doing a lot of stuff manually w/ websockets - the API (sshoverws) can do this but I like dialError
	defaultLogger.Info("Initializing websockets dialer from client")
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	dialer := websocket.Dialer{}
	defaultLogger.Info("Made websockets dialer and dialing")
	url := "ws://127.0.0.1:2223"
	conn, resp, err := dialer.Dial(url,nil)
	if err != nil {
		defaultLogger.Error("websocket.Dialier.Dial: Dial fail: " + err.Error())
		dialError(url, resp, err)
		return
	}
	defaultLogger.Info("Dialed no error, wrapping websockets connection")
	ioConn := sshoverws.WrapConn(conn)
	defer ioConn.Close()

	// Now Starting SSH
	defaultLogger.Info("Starting SSH server over wrapped websockets connection")
	sshConn, chans, reqs, err := ssh.NewServerConn(ioConn, config) // TODO: This isn't working
	if err != nil {
		defaultLogger.Error("NewServerConn error: " + err.Error())
		return
	}
	defaultLogger.Info("ssh.NewServerConn returned: " + sshConn.RemoteAddr().String() +" " + string(sshConn.ClientVersion() ) )
	// Discard all global out-of-band requests
	defaultLogger.Info("Running DiscardRequests as a goroutine")
	go ssh.DiscardRequests(reqs)
	// Accept all channels
	defaultLogger.Info("Running handleChannels as a goroutine")
	go handleChannels(chans)
	defaultLogger.Info("Calling sshConn.Wait()")
	err = sshConn.Wait()
	if err != nil {
		defaultLogger.Error("sshConn.Wait(): " + err.Error())
	}
}
