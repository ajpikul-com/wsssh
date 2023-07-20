package main

import (
	"flag"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh"
	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/mux"
)

var defaultLogger ilog.LoggerInterface
var HostPort string = "127.0.0.1:2223"

func init() {
	defaultLogger = &ilog.ZapWrap{Sugar: true}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	wsconn.SetDefaultLogger(defaultLogger)
}

func ServeWSConn(w http.ResponseWriter, r *http.Request) {
	defaultLogger.Info("Incoming: Req: " + r.Host + ", " + r.URL.Path)
	conn, err := sshoverws.Upgrade(w, r)
	if err != nil {
		defaultLogger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		defaultLogger.Info("Closing WSConn")
		if err := conn.Close(); err != nil {
			defaultLogger.Error(err.Error())
		}
	}()
}

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/", handleProxy)
	s := &http.Server{
		Addr:           HostPort,
		Handler:        m,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	defaultLogger.Info("Serving an http(s) server!")
	err := s.ListenAndServe()
	if err != nil {
		defaultLogger.Error("http.Server.ListenAndServe: " + err.Error())
	}
	defaultLogger.Info("Done with all on server")
}
