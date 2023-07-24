module github.com/ajpikul-com/wsssh/testclient

replace github.com/ajpikul-com/wsssh/wsconn => ../wsconn

go 1.20

require (
	github.com/ajpikul-com/ilog v0.0.0-20230714204235-1f6eb0175462
	github.com/ajpikul-com/wsssh/wsconn v0.0.0-20230721063641-93a47da82b32
	github.com/gorilla/websocket v1.5.0
	golang.org/x/crypto v0.11.0
)

require (
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)