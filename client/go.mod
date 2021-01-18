module github.com/ayjayt/AccessTunnel/client

go 1.14

require (
	github.com/ayjayt/AccessTunnel/sshoverws v0.0.0-20210118022152-bfdadb080504
	github.com/ayjayt/ilog v0.0.0-20210115032610-15372227e4a5
	github.com/gorilla/websocket v1.4.2
	github.com/kr/pty v1.1.8
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
)

replace golang.org/x/crypto => /home/ajp/gohack/golang.org/x/crypto
