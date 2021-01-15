module github.com/ayjayt/AccessTunnel/client

go 1.14

require (
	github.com/gorilla/websocket v1.4.2
	github.com/kr/pty v1.1.8
	github.com/ayjayt/AccessTunnel/sshoverws v0.0.0-20200614220045-064d68c1a884
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
)

replace golang.org/x/crypto => /home/ajp/gohack/golang.org/x/crypto

replace github.com/ayjayt/AccessTunnel/sshoverws => /home/ajp/gohack/github.com/ayjayt/AccessTunnel/sshoverws
