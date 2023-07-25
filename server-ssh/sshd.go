package main

import (
	"errors"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/ajpikul-com/wsssh/wsconn"
)

// Panicing probably isn't correct as server could be doing other things, and this is called per connection TODO
func GetServer(wsconn *wsconn.WSConn, clients string, privateKey string) (*ssh.ServerConn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {

	authorizedKeysBytes, err := os.ReadFile(clients)
	if err != nil {
		panic("Failed to load auth keys file " + err.Error())
	}
	authorizedKeysMap := map[string]string{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, comment, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			panic(err.Error())
		}
		authorizedKeysMap[string(pubKey.Marshal())] = comment
		authorizedKeysBytes = rest
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			defaultLogger.Debug("Running Public Key Callback")
			comment, ok := authorizedKeysMap[string(pubKey.Marshal())]
			if ok {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
						"comment":   comment,
					},
				}, nil
			}
			return nil, errors.New("No access for " + c.User())
		},
	}
	privateBytes, err := os.ReadFile(privateKey)
	if err != nil {
		panic("Problem loading private key file")
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Couldn't parse private key")
	}

	config.AddHostKey(private)

	return ssh.NewServerConn(wsconn, config)
}
