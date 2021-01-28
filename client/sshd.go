package main

import (
	"encoding/binary"
	"io"
	"os/exec"
	"syscall"
	//"sync"
	"strconv"
	"unsafe"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)
var channelID int = 0
func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		defaultLogger.Info("handleChannels: new Channel received")
		go handleChannel(newChannel, channelID)
		channelID += 1
	}
}

func handleChannel(newChannel ssh.NewChannel, channelID int) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type: " + t)
		defaultLogger.Info("Channel rejected by handleChannel: " + newChannel.ChannelType())
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	defaultLogger.Info("HandleChannels is accepting a channel")
	connection, requests, err := newChannel.Accept()
	if err != nil {
		defaultLogger.Error("ssh.NewChannel.Accept(): " + err.Error())
		return
	}

	// BASH
	defaultLogger.Info("Handle channel is setting up a bash command")
	bash := exec.Command("bash")

	// Prep Close
	close := func() {
		defaultLogger.Info("In close lambda and closing channel connection")
		connection.Close()
		defaultLogger.Info("Calling Bash.Process.Wait()") // TODO: Is this right? This should close bash.
		_, err := bash.Process.Wait()
		if err != nil {
			defaultLogger.Error("handleChannel/bash.Process.Wait:" + err.Error())
		}
		defaultLogger.Info("Session closed and bash too")
	}

	// Allocate a terminal for this channel
	defaultLogger.Info("Creating pty w/ bash attached")
	bashf, err := pty.Start(bash)
	if err != nil {
		defaultLogger.Error("pty.Start(): "+ err.Error())
		close()
		return
	}

	// pipe sessions to pty and vis-a-versa
	//var once sync.Once
	go func() {
		defaultLogger.Info("Copying channel to bash")
		io.Copy(connection, bashf) // This seems to receive EOF's but I'm not sure why or when. Certainly it shouldn't be done.
		defaultLogger.Info("Copied channel to bash closing go func channelID " + strconv.Itoa(channelID))
	//	once.Do(close)
	}()
	go func() {
		defaultLogger.Info("Copying bash to channel")
		io.Copy(bashf, connection)
		defaultLogger.Info("Copied bash to channel closing go func channelID " + strconv.Itoa(channelID))
	//	once.Do(close)
	}()

	for req := range requests {
		defaultLogger.Info("Request on a channel received: " + req.Type)
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

