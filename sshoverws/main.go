package sshoverws

import (
	"context"
	"flag"
	"log"

	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
)

var upgrader websocket.Upgrader

type WSAddress struct {
	host string
}
func (wsa *WSAddress) Network() string {
	return "ws"
}
func (wsa *WSAddress) String() string {
	return host
}
type WSTranport struct {
	conn websocket.Conn
}

func (wst *WSTransport) Read(b []byte) (n int, err error) {
	// Errors from here are fatal, connection must be reset
	tp, r, err := wst.conn.NextReader()
	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseAbnormalClosure,
	) {
		return
	}
	if err != nil {
		return 0, err
	}
	if mt != websocket.BinaryMessage {
		log.Fatal("Not a binary message")
	}
	return r.Read(b)
}
func (wst *WSTransport) Write(b []byte) (n int, err error) {
	wc, err := wst.conn.NextWriter(websocket.BinaryMessage)
	if (err != nil) {
		return 0, err
	}
	n, err := wc.Write(b)
	if err != nil {
		return n, err
	}
	err = wc.Close()
	return n, err
}
func (wst *WSTransport) Close() error {
	return wst.conn.Close()
}
func (wst *WSTransport) LocalAddr() Addr {
	// TODO UNIMPLEMENTED
	return &wsAddres{host:"example"}
}
func (wst *WSTransport) RemoteAddr() Addr {
	// TODO UNIMPLEMENTED
	return &WSAddress{host:"example"}
}
func (wst *WSTransport) SetDeadline(t time.Time) error {
	// TODO UNIMPLEMENTED
	return nil
}
func (wst *WSTransport) SetReadDeadline(t time.Time) error {
	// TODO UNIMPLEMENTED
	return nil
}
func (wst *WSTransport) SetWriteDeadline(t time.Time) error {
	// TODO UNIMPLEMENTED
	return nil
}

func Upgrade(w http.ResponseWriter, r *http.Request) *WSTransport, error {
	conn, err := upgrade.Upgrade(w, r, nil)
	return &WSTransport{conn:conn}, err
}

