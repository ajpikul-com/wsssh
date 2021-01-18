package main

import (
	"time"
	"net"
	"net/http"
	"flag"

	"golang.org/x/crypto/ssh"

	"github.com/ayjayt/ilog"
	"github.com/ayjayt/AccessTunnel/sshoverws"
	"github.com/gorilla/mux"
)

var hostPrivateKey = flag.String("hostkey", "", "Path to the your private key")

var defaultLogger ilog.LoggerInterface


func init(){
	defaultLogger = new(ilog.SimpleLogger)
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	sshoverws.SetDefaultLogger(defaultLogger)
	flag.Parse()
	if (len(*hostPrivateKey) == 0) {
		defaultLogger.Error("server main.go flags: No host private key set")
		defaultLogger.Info("Skipping above error since no auth is implemented yet")
	}
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	defaultLogger.Info("Req: " + r.Host +", " + r.URL.Path) // TODO: So I guess we're taking all paths here?
	conn, err := sshoverws.Upgrade(w, r) // so it's not a handler
	if err != nil {
		defaultLogger.Error("AccessTunnel/server/main.go handleProxy sshoverws.Upgrade err: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - sumtin broke")) // TODO: not sure if Upgrade writes a response
		return
	}
	defaultLogger.Info("Setting up new ClientConn to the edge device")
	sshClientConn, chans, reqs, err := ssh.NewClientConn(conn, "device", &ssh.ClientConfig{
		User: "ajp",
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// key is the hosts public key, which has already been certified
			defaultLogger.Info("Hostname: " + hostname + ", remote: " + remote.Network())
			return nil
		},
	})
	defer conn.Close()
	if err != nil {
		defaultLogger.Error("AccessTunnel/server/main.go handleProxy ssh.NewClientConn err: " + err.Error())
		return // TODO: A way to close conn w/ message? Multiplex over one message. 
					 // Client COULD be set to accept text... not binary... as well. // Assumes websockets is up
	}
	defaultLogger.Info("Setting up new client")
	sshClient := ssh.NewClient(sshClientConn, chans, reqs)
	defaultLogger.Info("Start Session") // Set's up one sessions
	session, err := sshClient.NewSession()
	if err != nil {
		defaultLogger.Error("AccessTunnel/server/main.go sshClient.NewSession() err: " + err.Error())
		return
	}
	defer session.Close()
	// Do I need to request psuedoterminal? 
	// What does this do without it?
	defaultLogger.Info("Calling Shell")
	err = session.Shell()
	if err != nil {
		defaultLogger.Error("AccessTunnel/server/main.go sshClient.Session.Shell() err: " + err.Error())
		return
	}
	defaultLogger.Info("Shell up")
	session.Wait()
}

func main(){
	m := mux.NewRouter()
	m.HandleFunc("/", handleProxy)
	s := &http.Server {
		Addr: "127.0.0.1:2223",
		Handler:	m,
		ReadTimeout:	10 * time.Second,
		WriteTimeout:	10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	defaultLogger.Info("Serving!")
	err := s.ListenAndServe()
	defaultLogger.Error("AccessTunnel/server/main.go main http.Server.ListenAndServer err: " + err.Error())
}

