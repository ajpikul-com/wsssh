module github.com/ayjayt/AccessTunnel/client

go 1.14

require (
	github.com/ayjayt/AccessTunnel/sshoverws v0.0.0-20210120024753-88fee875e242
	github.com/ayjayt/ilog v0.0.0-20210115032610-15372227e4a5
	github.com/creack/pty v1.1.11 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/kr/pty v1.1.8
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
)

replace golang.org/x/crypto => /home/ajp/gohack/golang.org/x/crypto
