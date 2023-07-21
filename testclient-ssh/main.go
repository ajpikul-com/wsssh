package main

import (
	"bytes"
	"context"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ajpikul-com/wsssh/wsconn"
	gws "github.com/gorilla/websocket"
)

var url string = "ws://127.0.0.1:4648"

func ReadTexts(conn *wsconn.WSConn) {
	textChan := make(chan int)
	conn.TextChan = textChan
	p := make([]byte, 1024)
	for _ = range textChan {
		for conn.TextBuffer.Len() > 0 {
			n, err := conn.TextBuffer.Read(p)
			defaultLogger.Info("ReadTexts: " + string(p[0:n]))
			defaultLogger.Info("ReadTexts Remaining: " + strconv.Itoa(conn.TextBuffer.Len()))
			if err != nil {
				defaultLogger.Error("ReadTexts: TextBuffer.Read():" + err.Error())
				break
			}
		}
	}
	defaultLogger.Info("ReadTexts Channel Closed")
	// The channel has been closed by someone else
}

func WriteText(conn *wsconn.WSConn) {
	for {
		_, err := conn.WriteText([]byte("Test Message")) // TODO Can we be sure this will write everything
		if err != nil {
			defaultLogger.Error("WriteText: wsconn.WriteText(): " + err.Error())
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func Pinger(conn *wsconn.WSConn) {
	for {
		err := conn.WritePing([]byte("Perring"))
		if err != nil {
			defaultLogger.Error("Pinger: " + err.Error())
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	// Doing a lot of stuff manually w/ websockets - still not sure why I'm doing it this way
	defaultLogger.Info("Initializing websockets dialer from client")

	_, cancel := context.WithCancel(context.Background())
	defer cancel() // Is this really necessary?
	dialer := gws.Dialer{}
	conn1, resp, err := dialer.Dial(url, nil)
	if err != nil {
		defaultLogger.Error("websocket.Dialier.Dial: Dial fail: " + err.Error())
		dumpResponse(resp)
		return
	}
	wssshConn, err := wsconn.New(conn1)
	if err != nil {
		panic(err.Error())
	}

	go ReadTexts(wssshConn)

	hkc := func(h string, remote net.Addr, key ssh.PublicKey) error {
		defaultLogger.Info("HKC")
		defaultLogger.Info(h)
		defaultLogger.Info(remote.String())
		defaultLogger.Info(key.Type())
		return nil
	}
	privateBytes, err := os.ReadFile("/home/ajp/.ssh/id_ed25519")
	if err != nil {
		panic("what happened to our private key")
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("couldn't parse private key")
	}

	config := &ssh.ClientConfig{
		User: "ajp",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(private),
		},
		HostKeyCallback: hkc,
	}

	conn, chans, reqs, err := ssh.NewClientConn(wssshConn, "127.0.0.1:4648", config)
	if err != nil {
		defaultLogger.Error("We given up??" + err.Error())

		time.Sleep(100 * time.Second)
		panic("sleepy")
	}
	defaultLogger.Info("We in bb")
	client := ssh.NewClient(conn, chans, reqs)
	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	defaultLogger.Info("We a client")
	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()
	defaultLogger.Info("Got a session!")
	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/whoami"); err != nil {
		defaultLogger.Error("Failed to run: " + err.Error())
	}
	defaultLogger.Info("Ran a session!")
	defaultLogger.Info(b.String())
	// ssh calls closed for us
	time.Sleep(10 * time.Second)
	wssshConn.Conn.WriteControl(gws.CloseMessage, []byte(""), time.Time{})
	err = wssshConn.Close()
	if err != nil {
		defaultLogger.Info("Tried to close: " + err.Error())
	}
}
