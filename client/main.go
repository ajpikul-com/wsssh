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
				defaultLogger.Error("Failed to read HTTP body: "+ err.Error())
				// return ?
			}
			extra = "Body:\n" + string(b)
		defaultLogger.Info(err.Error() + ": HTTP error: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status + "\n" + extra)

	}
	defaultLogger.Error("Dial to " + url + " fail: " + err.Error())
}

func main() {
	config := &ssh.ServerConfig{
		NoClientAuth: true, // TODO lol what but wait
	}

	privateBytes, err := ioutil.ReadFile("/home/ajp/.ssh/id_ed25519")
	if err != nil {
		defaultLogger.Error("You must set a proper hostkey with -hostkey")
		return
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		defaultLogger.Error("Failed to parse private key")
		return
	}

	config.AddHostKey(private)


	// Doing a lot of stuff manually w/ websockets - the API (sshoverws) can do this but I like dialError
	defaultLogger.Info("Starting dialer")
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	dialer := websocket.Dialer{}
	defaultLogger.Info("Made dialer")
	conn, resp, err := dialer.Dial("ws://houston.osoximeter.com:2223",nil)
	if err != nil {
		defaultLogger.Error("Err: " + err.Error())
	}
	defaultLogger.Error("Dialed no error")
	if err != nil {
		dialError("houston.osoximeter.com", resp, err)
	}
	ioConn := sshoverws.WrapConn(conn)
	defer ioConn.Close()

	// Now Starting SSH
	defaultLogger.Error("Starting ssh")
	sshConn, chans, reqs, err := ssh.NewServerConn(ioConn, config)
	if err != nil {
		defaultLogger.Error("Failed to handshake " + err.Error())
		return
	}

	defaultLogger.Info("New SSH cnx from " + sshConn.RemoteAddr().String() +" " + string(sshConn.ClientVersion() ) )
		// Discard all global out-of-band requests
	go ssh.DiscardRequests(reqs)
		// Accept all channels
	go handleChannels(chans)
	sshConn.Wait()
}
