module github.com/ayjayt/AccessTunnel/client

go 1.14

require (
	github.com/ayjayt/AccessTunnel/sshoverws v0.0.0-20210129024551-a71661b22d6b
	github.com/ayjayt/ilog v0.0.0-20210115032610-15372227e4a5
	github.com/creack/pty v1.1.11 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/kr/pty v1.1.8
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
)

replace github.com/ayjayt/AccessTunnel/sshoverws => ../sshoverws
