package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/mux"
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
	conn, err := wsconn.Upgrade(w, r)
	if err != nil {
		defaultLogger.Error("Server: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	readBuffer := make([]byte, 1024)
	for n, err := conn.Read(readBuffer); err == nil; n, err = conn.Read(readBuffer) {
		defaultLogger.Info("Server:")
		defaultLogger.Info("Server: In read:")
		defaultLogger.Info("Server: N is: " + strconv.Itoa(n))
		defaultLogger.Info("Server: " + string(readBuffer[:]))
		defaultLogger.Info("Server:")
	}

	defer func() {
		defaultLogger.Info("Server: Closing WSConn")
		if err := conn.Close(); err != nil {
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
