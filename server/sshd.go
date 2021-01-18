package main

import (
	"io/ioutil"
	"log"
	"net"
	"fmt"
	"encoding/binary"
	"io"
	"os/exec"
	"syscall"
	"sync"
	"unsafe"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

func runSSHD(port string) {
	panic("Asbolutely do not do this")
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	privateBytes, err := ioutil.ReadFile(*hostPrivateKey)
	if err != nil {
		log.Fatalf("You must set a proper hostkey with -hostkey")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key")
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Couldn't open port %s: %s", port, err)
	}

	log.Printf("Opening port %s", port)
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept a request (%s)", err)
			continue
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH cnx from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

		// Discard all global out-of-band requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go handleChannels(chans)
	}

}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	// BASH
	bash := exec.Command("bash")

	// Prep Close
	close := func() {
		connection.Close()
		_, err := bash.Process.Wait()
		if err != nil {
			log.Printf("Failed to exit bash (%s)", err)
		}
		log.Printf("Session closed")
	}

	// Allocate a terminal for this channel
	log.Print("Creating pty...")
	bashf, err := pty.Start(bash)
	if err != nil {
		log.Printf("Could not start pty (%s)", err)
		close()
		return
	}

	// pipe sessions to pty and vis-a-versa
	var once sync.Once
	go func() {
		io.Copy(connection, bashf) // so this is uh..  the right message type
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, connection)
		once.Do(close)
	}()

	go func() {
		for req := range requests { // what if it requests a new channel? // Can't request channels on channels
			switch req.Type {
			case "shell": // Subsystem?
				// We only accept default shell, no command in payload
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req": // this is where we start bash
				termLen := req.Payload[3]
				w, h:= parseDims(req.Payload[termLen+4:])
				SetWinsize(bashf.Fd(), w, h)
				req.Reply(true, nil)
			case "window-change":
				w, h := parseDims(req.Payload)
				SetWinsize(bashf.Fd(), w, h)
			}
		}
	}()
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height	uint16
	Width	uint16
	x	uint16 // unused
	y	uint16 // unused
}

// SetWinsize sets the size of the given pty.
func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}

