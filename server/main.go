package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var defaultLogger ilog.LoggerInterface
var HostPort string = "127.0.0.1:2223"

func init() {
	defaultLogger = &ilog.SimpleLogger{}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	wsconn.SetDefaultLogger(defaultLogger)
}

func ServeWSConn(w http.ResponseWriter, r *http.Request) {
	defaultLogger.Info("Server: Incoming Req: " + r.Host + ", " + r.URL.Path)
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		defaultLogger.Error("Server: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	wsconn, err := wsconn.New(conn)
	if err != nil {
		panic(err.Error())
	}

	// TODO, SPIN THESE THINGS OFF INTO GOROUTINES
	var n int = 0
	readBuffer := make([]byte, 1024)
	for err == nil {
		defaultLogger.Info("Server: Starting an inner read")
		for n, err = wsconn.Read(readBuffer); n != 0; n, err = wsconn.Read(readBuffer) {
			defaultLogger.Info("Server:")
			defaultLogger.Info("Server: In read:")
			defaultLogger.Info("Server: N is: " + strconv.Itoa(n))
			defaultLogger.Info("Server: " + string(readBuffer[0:n])) // this will stop output on on ascii character, but we should use buffer length
			if err != nil {                                          // here we also break on error, couldn't put it in 4 because
				// I wanted to see the buffer first
				defaultLogger.Error("Server: BREAK INNER READ Err:" + err.Error())
				defaultLogger.Info("Server:")
				break
			}
		}
	}

	defer func() {
		defaultLogger.Info("Server: Closing WSConn")
		if err := wsconn.Close(); err != nil {
			defaultLogger.Error("Server: " + err.Error())
		}
	}()
}

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/", ServeWSConn)
	s := &http.Server{
		Addr:           HostPort,
		Handler:        m,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	defaultLogger.Info("Server: Initiating server")
	err := s.ListenAndServe()
	if err != nil {
		defaultLogger.Error("Server: http.Server.ListenAndServe: " + err.Error())
	}
	defaultLogger.Info("Server exiting")
}
