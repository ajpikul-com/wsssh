package main

import (
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var defaultLogger ilog.LoggerInterface
var HostPort string = "127.0.0.1:4648"

func init() {
	defaultLogger = &ilog.SimpleLogger{}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	packageLogger := &ilog.SimpleLogger{}
	packageLogger.Level(ilog.INFO)
	packageLogger.Init()
	wsconn.SetDefaultLogger(packageLogger)
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

func ReadTexts(conn *wsconn.WSConn) {
	defaultLogger.Debug("Starting to read texts")
	channel, _ := conn.SubscribeToTexts()
	for s := range channel {
		defaultLogger.Info("ReadTexts: " + s)
	}
	defaultLogger.Debug("ReadTexts Channel Closed")
	// The channel has been closed by someone else
}

func ServeWSConn(w http.ResponseWriter, r *http.Request) {
	defaultLogger.Debug("Server: Incoming Req: " + r.Host + ", " + r.URL.Path)
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	},
	}
	gorrilaconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		defaultLogger.Error("Upgrade: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	wsconn, err := wsconn.New(gorrilaconn)
	if err != nil {
		panic(err.Error()) // This actually is very severe
	}
	defer func() {
		defaultLogger.Debug("Closing WSConn")
		// Doesn't warn client, just closes
		if err := wsconn.CloseAll(); err != nil {
			defaultLogger.Error("wsconn.CloseAll(): " + err.Error())
		}
	}()

	sshconn, chans, reqs, err := GetServer(wsconn, "/home/ajp/systems/public_keys/ajp", "/home/ajp/.ssh/id_ed25519")
	if err != nil {
		defaultLogger.Error("GetServer(): " + err.Error())
		return
	}
	defaultLogger.Info("Welcome, " + sshconn.Permissions.Extensions["comment"])
	go ReadTexts(wsconn)
	go ssh.DiscardRequests(reqs)
	for _ = range chans {
		// We do nothing for you
		// But it keeps the connection open: i suppose we wait for the client to close?
		// Otherwise it just stays open
	}
	defaultLogger.Info(sshconn.Permissions.Extensions["comment"] + " disconnected")
}

func main() {
	m := mux.NewRouter()
	// Add two gets here. Why are we using mux? I don't know
	m.HandleFunc("/", ServeWSConn)
	s := &http.Server{
		Addr:           HostPort,
		Handler:        m,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	defaultLogger.Info("Initiating server")
	err := s.ListenAndServe()
	if err != nil {
		defaultLogger.Error("http.Server.ListenAndServe: " + err.Error())
	}
	defaultLogger.Info("Server exiting")
}
