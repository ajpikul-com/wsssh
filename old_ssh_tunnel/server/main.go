package main

import (
	"time"
	"net"
	"net/http"
	"flag"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/ayjayt/ilog"
	"github.com/ayjayt/AccessTunnel/sshoverws"
	"github.com/gorilla/mux"
)

var hostPrivateKey = flag.String("hostkey", "", "Path to the your private key")

var defaultLogger ilog.LoggerInterface


func init(){
	defaultLogger = new(ilog.ZapWrap)
	defaultLogger.(*ilog.ZapWrap).Paths = []string{"./server.log"}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	sshoverws.SetDefaultLogger(defaultLogger)
	flag.Parse()
	if (len(*hostPrivateKey) == 0) {
		defaultLogger.Error("Server main.go flags: No ssh-local private key set")
		defaultLogger.Info("Skipping private key error since no auth is implemented yet")
	}
}


func handleProxy(w http.ResponseWriter, r *http.Request) {
	defaultLogger.Info("Connection made to handleProxy: Req: " + r.Host +", " + r.URL.Path)
	conn, err := sshoverws.Upgrade(w, r)
	if err != nil {
		defaultLogger.Error("handleProxy sshoverws.Upgrade: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - sumtin broke"))
		return
	}
	defer func() {
		defaultLogger.Info("Closing sshoverws connection.")
		if err := conn.Close(); err != nil {
			defaultLogger.Error("sshoverws.WSTransport.Close(): " + err.Error())
		}
	}()

	if err = conn.WriteText("Test Text"); err != nil {
		defaultLogger.Error("WriteText(): " + err.Error())
	}

	// Start the client connection- this is where we identify the client
	// This is where the globals have been set up.
	defaultLogger.Info("Creating an ssh.NewClientConn to the edge device")
	sshClientConn, chans, reqs, err := ssh.NewClientConn(conn, r.RemoteAddr, &ssh.ClientConfig{
	//sshClientConn, _, _, err := ssh.NewClientConn(conn, r.RemoteAddr, &ssh.ClientConfig{
		User: "ajp",
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// the hosts public key was traded by diffee hellman, we've verified that it has the private key
			// couldn't it just use the public key and then trade a nonce?
			defaultLogger.Info("HostkeyCallback: Hostname: " + hostname + ", remote: " + remote.Network())
			return nil
		},
	})
	if err != nil {
		defaultLogger.Error("ssh.NewClientConn err: " + err.Error())
		return
	}
	defer func() {
		defaultLogger.Info("Closing an ssh.ClientConn")
		if err := sshClientConn.Close(); err != nil {
			defaultLogger.Error("sshClientConn.Close: " + err.Error())
		}
	}()
	defaultLogger.Info("Setting up new client based on a ssh.ClientConn")
	sshClient := ssh.NewClient(sshClientConn, chans, reqs)

	// Now we're asking the client for individual channels. We've already asked for a particular user though. Can we multiplex multple NewClientConns over the same websockets transport?
	// Also, this sets up streams
	defaultLogger.Info("Starting a new ssh session")
	session, err := sshClient.NewSession()
	if err != nil {
		defaultLogger.Error("sshClient.NewSession(): " + err.Error())
		return
	}
	defer session.Close()
	// Needs to set up pipes- do these work?
	session.Stdin = os.Stdin
	session.Stderr = os.Stderr
	session.Stdout = os.Stdout

	// All of this basically does reqs
	defaultLogger.Info("Calling Shell on an ssh session")
	err = session.Shell()
	if err != nil {
		defaultLogger.Error("sshClient.Session.Shell(): " + err.Error())
		return
	}

	// Waiting just stays here and keeps the connection open. But basically we want to keep this connection open continuously.
	defaultLogger.Info("Setting wait on an ssh session")
	err = session.Wait()
	if err != nil {
		defaultLogger.Error("session.Wait(): " + err.Error())
	}
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
	defaultLogger.Info("Serving an http(s) server!")
	err := s.ListenAndServe()
	if err != nil {
		defaultLogger.Error("http.Server.ListenAndServe: " + err.Error())
	}
	defaultLogger.Info("Done with all on server")
}

