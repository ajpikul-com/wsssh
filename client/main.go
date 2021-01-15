package main

import (
	"context"
	"log"
	"io/ioutil"
	"net/http"
	"golang.org/x/crypto/ssh"

	"github.com/gorilla/websocket"
	"github.com/ayjayt/AccessTunnel/sshoverws"
)

func dialError(url string, resp *http.Response, err error) {
	if resp != nil {
		extra := ""
		if true {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read HTTP body: %v", err)
			}
			extra = "Body:\n" + string(b)
		}
		log.Printf("%s: HTTP error: %d %s\n%s", err, resp.StatusCode, resp.Status, extra)

	}
	log.Printf("Dial to %q fail: %v", url, err)
}

func main() {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	privateBytes, err := ioutil.ReadFile("/home/ajp/.ssh/id_ed25519")
	if err != nil {
		log.Fatalf("You must set a proper hostkey with -hostkey")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key")
	}

	config.AddHostKey(private)

	log.Printf("Starting dialer")
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	dialer := websocket.Dialer{}
	log.Printf("Made dialer")
	conn, resp, err := dialer.Dial("ws://houston.osoximeter.com",nil)
	if err != nil {
		log.Printf("Err (%s)", err)
	}
	log.Printf("Dialed no error")
	if err != nil {
		dialError("houston.osoximeter.com", resp, err)
	}

	ioConn := sshoverws.WrapConn(conn)

	defer ioConn.Close()
	log.Printf("Starting ssh")
	sshConn, chans, reqs, err := ssh.NewServerConn(ioConn, config)
	if err != nil {
		log.Printf("Failed to handshake (%s)", err)
		return
	}

	log.Printf("New SSH cnx from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band requests
	go ssh.DiscardRequests(reqs)
		// Accept all channels
	go handleChannels(chans)
	sshConn.Wait()
}
