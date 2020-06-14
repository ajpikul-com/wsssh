package sshoverws

import (
	"log"
	"net"
	"time"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader

type WSAddress struct {
	host string
}
func (wsa *WSAddress) Network() string {
	return "ws"
}
func (wsa *WSAddress) String() string {
	return wsa.host
}
type WSTransport struct {
	conn *websocket.Conn
}

func (wst *WSTransport) Read(b []byte) (n int, err error) {
	// Errors from here are fatal, connection must be reset
	mt, r, err := wst.conn.NextReader()
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
	n, err = wc.Write(b)
	if err != nil {
		return n, err
	}
	err = wc.Close()
	return n, err
}
func (wst *WSTransport) Close() error {
	return wst.conn.Close()
}
func (wst *WSTransport) LocalAddr() net.Addr {
	// TODO UNIMPLEMENTED
	return &WSAddress{host:"example"}
}
func (wst *WSTransport) RemoteAddr() net.Addr {
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

func Upgrade(w http.ResponseWriter, r *http.Request) (*WSTransport, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	return &WSTransport{conn:conn}, err
}

