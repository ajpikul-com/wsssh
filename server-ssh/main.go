package main

import (
	"errors"
	"net/http"
	"os"
	//	"os/exec"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var defaultLogger ilog.LoggerInterface
var HostPort string = ":4648"

func init() {
	defaultLogger = &ilog.SimpleLogger{}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	wsconn.SetDefaultLogger(defaultLogger)
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

func ServeWSConn(w http.ResponseWriter, r *http.Request) {
	defaultLogger.Info("Server: Incoming Req: " + r.Host + ", " + r.URL.Path)
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	},
	}
	conn1, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		defaultLogger.Error("Server: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	wsconn, err := wsconn.New(conn1)
	if err != nil {
		panic(err.Error())
	}

	go ReadTexts(wsconn)

	/// EASY CHEESEY

	authorizedKeysBytes, err := os.ReadFile("/home/ajp/systems/public_keys/ajp")
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
	defaultLogger.Info("Public keys:")
	for k, v := range authorizedKeysMap {
		defaultLogger.Info(k)
		defaultLogger.Info(strconv.FormatBool(v))
	} // GIBERISH
	/// SEEMS IMPORTANT

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
	privateBytes, err := os.ReadFile("/home/ajp/.ssh/id_ed25519")
	if err != nil {
		panic("what happened to our private key")
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("couldn't parse private key")
	}

	config.AddHostKey(private)

	conn, chans, reqs, err := ssh.NewServerConn(wsconn, config)
	if err != nil {
		panic("Couldn't connect to conn " + err.Error())
	}
	defaultLogger.Info("Logged in with " + conn.Permissions.Extensions["pubkey-fp"])

	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		defaultLogger.Info("Found a chan")
		if newChannel.ChannelType() != "session" {
			defaultLogger.Info("Not a session")
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		defaultLogger.Info("Accepted the session in the channel")
		if err != nil {
			panic("Channel accept failed " + err.Error())
		}
		go func(in <-chan *ssh.Request) {
			for req := range in {
				defaultLogger.Info("Processing a request " + req.Type)
				if req.WantReply {
					defaultLogger.Info("Wants a reply")
					req.Reply(true, nil)
				}
				defaultLogger.Info("Req: " + string(req.Payload[:]))
				defaultLogger.Info("Now reading input to channel")
				defaultLogger.Info("trying to write channel")
				//p := make([]byte, 1024)
				//_, err := channel.Read(p)
				_, err := channel.Write([]byte("He who eats kitties"))
				defaultLogger.Info("wrote channel")
				if err != nil {
					defaultLogger.Info("Channel Read: " + err.Error())
					break
				}
				channel.SendRequest("exit-status", false, []byte{0b00, 0b00, 0b00, 0b00})
				time.Sleep(5 * time.Second)
				defaultLogger.Info("Req reply and channel close")
				time.Sleep(5 * time.Second)
				channel.Close()
			}
		}(requests)

	}
	defaultLogger.Info("Seems like we're closing the function main")
	time.Sleep(5 * time.Second)
	defer func() {
		defaultLogger.Info("Server: Closing WSConn")
		if err := wsconn.Close(); err != nil {
			defaultLogger.Error("Server: " + err.Error())
		}
	}()
}

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/", ServeWSConn)
	s := &http.Server{
		Addr:           HostPort,
		Handler:        m,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	defaultLogger.Info("Server: Initiating server")
	err := s.ListenAndServe()
	if err != nil {
		defaultLogger.Error("Server: http.Server.ListenAndServe: " + err.Error())
	}
	defaultLogger.Info("Server exiting")
}
