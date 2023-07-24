package main

import (
	"errors"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/ajpikul-com/wsssh/wsconn"
)

// Panicing probably isn't correct as server could be doing other things, and this is called per connection TODO
func GetServer(wsconn *wsconn.WSConn, clients string, privateKey string) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {

	authorizedKeysBytes, err := os.ReadFile(clients)
	if err != nil {
		panic("Failed to load auth keys file " + err.Error())
	}
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			panic(err.Error())
		}
		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			defaultLogger.Info("Running Public Key Callback")
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, errors.New("No access for " + c.User())
		},
	}
	privateBytes, err := os.ReadFile(privateKey)
	if err != nil {
		panic("what happened to our private key")
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("couldn't parse private key")
	}

	config.AddHostKey(private)

	return ssh.NewServerConn(wsconn, config)
}
