package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/websocket"
)

var defaultLogger ilog.LoggerInterface

func init() {
	defaultLogger = &ilog.ZapWrap{Sugar: true}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	wsconn.SetDefaultLogger(defaultLogger)
}

func dumpResponse(resp *http.Response) {
	if resp != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			defaultLogger.Error(err.Error())
			// return ?
		}
		extra := "Body:\n" + string(b)
		defaultLogger.Error("HTTP Response Info: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status + "\n" + extra)
	}
}

var url string = "ws://127.0.0.1:2223" // Do I really need the scheme TODO

func main() {
	// Doing a lot of stuff manually w/ websockets - the API (sshoverws) can do this but I like dialError
	defaultLogger.Info("Initializing websockets dialer from client")

	_, cancel := context.WithCancel(context.Background())
	defer cancel() // Is this really necessary?
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(url, nil)
	if err != nil {
		defaultLogger.Error("websocket.Dialier.Dial: Dial fail: " + err.Error())
		dumpResponse(resp)
		return
	}
	wssshConn := wsconn.WrapConn(conn)
	defer wssshConn.Close()
	_ = wssshConn

	// Not we have a wrapped connection, we need to be able to call things on it.
}
