package main

import (
	"golang.org/x/crypto/ssh"
	"os"

	"github.com/ajpikul-com/wsssh/wsconn"
)

// TODO user arguments not static strings
func GetClient(conn *wsconn.WSConn, username string, myPrivateKey string, remotePublicKey string) (*ssh.Client, error) {

	// add public key read here TODO
	publicBytes, err := os.ReadFile(remotePublicKey)
	if err != nil {
		return nil, err
	}
	public, _, _, _, err := ssh.ParseAuthorizedKey(publicBytes)
	defaultLogger.Debug("Public Key:" + string(publicBytes))
	hostKeyCallback := ssh.FixedHostKey(public)

	privateBytes, err := os.ReadFile(myPrivateKey)
	if err != nil {
		return nil, err
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(private),
		},
		HostKeyCallback: hostKeyCallback,
	}

	sshconn, chans, reqs, err := ssh.NewClientConn(conn, "", config)
	if err != nil {
		return nil, err
	}
	client := ssh.NewClient(sshconn, chans, reqs)
	return client, nil
}

// NOTE: client's have sessions, which are a great way to use feature of ssh.
// Client close close the underlying session, but don't send message over ssh to server to close. It seems ssh stays reading.
