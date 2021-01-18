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
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type: " + t)
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		defaultLogger.Error("Could not accept channel: "+ err.Error())
		return
	}

	// BASH
	defaultLogger.Info("Executing bash")
	bash := exec.Command("bash")

	// Prep Close
	close := func() {
		connection.Close()
		_, err := bash.Process.Wait()
		if err != nil {
			defaultLogger.Error("Failed to exit bash (" + err.Error() + ")")
		}
		defaultLogger.Info("Session closed")
	}

	// Allocate a terminal for this channel
	defaultLogger.Info("Creating pty...")
	bashf, err := pty.Start(bash)
	if err != nil {
		defaultLogger.Error("Could not start pty ("+ err.Error()+")")
		close()
		return
	}

	// pipe sessions to pty and vis-a-versa
	var once sync.Once
	go func() {
		io.Copy(connection, bashf)
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, connection)
		once.Do(close)
	}()

	go func() {
		for req := range requests {
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

