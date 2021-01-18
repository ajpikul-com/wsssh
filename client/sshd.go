package main

import (
	"encoding/binary"
	"io"
	"os/exec"
	"syscall"
	"sync"
	"unsafe"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

func handleChannels(chans <-chan ssh.NewChannel) {
	defaultLogger.Info("INFO: handleChannels: new Channel received")
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type: " + t)
		defaultLogger.Info("INFO: Channel rejected: " + newChannel.ChannelType())
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	defaultLogger.Info("INFO: Accepting a channel")
	connection, requests, err := newChannel.Accept()
	if err != nil {
		defaultLogger.Error("AccessTunnel/client/sshd.go ssh.NewChannel.Accept(): " + err.Error())
		return
	}

	// BASH
	defaultLogger.Info("INFO: Executing bash")
	bash := exec.Command("bash")

	// Prep Close
	close := func() {
		defaultLogger.Info("INFO: In close lambda and closing channel connection")
		connection.Close()
		defaultLogger.Info("INFO: Bash.Process.Wait()")
		_, err := bash.Process.Wait()
		if err != nil {
			defaultLogger.Error("AccessTunnel/client/sshd.org handleChannel/bash.Process.Wait:" + err.Error())
		}
		defaultLogger.Info("INFO: Session closed")
	}

	// Allocate a terminal for this channel
	defaultLogger.Info("INFO: Creating pty...")
	bashf, err := pty.Start(bash)
	if err != nil {
		defaultLogger.Error("AccessTunnel/client/sshd.go pty.Start(): "+ err.Error())
		close()
		return
	}

	// pipe sessions to pty and vis-a-versa
	var once sync.Once
	go func() {
		defaultLogger.Info("INFO: Copying channel to bash")
		io.Copy(connection, bashf)
		defaultLogger.Info("INFO: Copied channel to bash")
		once.Do(close)
	}()
	go func() {
		defaultLogger.Info("INFO: Copying bash to channel")
		io.Copy(bashf, connection)
		defaultLogger.Info("INFO: Copied bash to channel")
		once.Do(close)
	}()

	go func() {
		for req := range requests {
			defaultLogger.Info("INFO: Request received: " + req.Type)
			switch req.Type {
			case "shell":
				// We only accept default shell, no command in payload
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
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
	Width		uint16
	x				uint16 // unused
	y				uint16 // unused
}

// SetWinsize sets the size of the given pty.
func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}

