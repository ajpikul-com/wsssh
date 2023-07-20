package wsconn

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ajpikul-com/ilog"

	"github.com/gorilla/websocket"
)

var defaultLogger ilog.LoggerInterface

func init() {
	if defaultLogger == nil {
		defaultLogger = new(ilog.EmptyLogger)
	}
}

// SetDefaultLogger attaches a logger to the lib. See github.com/ajpikul-com/ilog
func SetDefaultLogger(newLogger ilog.LoggerInterface) {
	defaultLogger = newLogger
	defaultLogger.Info("Default Logger Set")
}

// WSConn satisfies the net.Conn interface
type WSConn struct {
	// The underlying websocket connection.
	Conn *websocket.Conn

	// READING STUFF
	// Effectively indicate if we're mid frame
	r  io.Reader
	mt int
	// TextBuffer
	TextBuffer *bytes.Buffer // We're going to have to initialize this TODO

	// WRITING STUFF
	writeMutex sync.Mutex
}

// New returns an initialized *WSConn
func New(conn *websocket.Conn) (wsconn *WSConn, err error) {
	wsconn = &WSConn{Conn: conn, r: nil, mt: 0, TextBuffer: new(bytes.Buffer)}
	wsconn.Conn.SetPingHandler(func(message string) error {
		err = wsconn.WritePong([]byte(message))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			// I don't know if we can do this, as we can't see Gorillaz net.Error struct
			return nil
		}
		return err
	})
	return wsconn, nil
}

// The main Read() function
// Only one thread can call read, it will be whatever is using WSConn like net.Conn
// It will be populating TextBuffer
// How do we read TextBuffer? I don't know. TODO
func (conn *WSConn) Read(b []byte) (n int, err error) {
	for { // We'll read until our first binary message
		if conn.mt == 0 { // Need to read until EOF before calling NextReader() again
			// Errors from here are fatal, connection must be reset
			mt, r, err := conn.Conn.NextReader()
			if err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseAbnormalClosure,
				) {
					defaultLogger.Info("wsconn.Read().NextReader(): A Close Return: " + err.Error())
					return 0, err
				}
				defaultLogger.Error("wsconn.Read().NextReader(): " + err.Error())
				return 0, err
			}

			conn.mt = mt
			conn.r = r

			if conn.mt > websocket.BinaryMessage {
				defaultLogger.Error("wsconn.Read(): unexpected message type")
				// Controls shouldn't get here
				return 0, errors.New("wsconn.Read(): Wrong error type received")
			}
		}
		for { // keep on reading until we return or break out
			n, err = conn.r.Read(b)
			if err != nil || b == nil {
				conn.r = nil
				conn.mt = 0
				if err == io.EOF {
					defaultLogger.Info("wsconn.Read(): reached EOF")
					err = nil
					break // break out of this forloop and go get a new reader
				} else {
					defaultLogger.Error("wsconn.Read(): NextReader's io.Reader.Read() " + err.Error())
					return n, err
				}
			}
			if conn.mt == websocket.TextMessage {
				conn.TextBuffer.Write(b) // will this really work? Do we need to indicare length?
			} else {
				return n, err
			}
		}
	}
}

// Write is the Write that net.Conn expects
func (conn *WSConn) Write(b []byte) (n int, err error) {
	return conn.write(b, websocket.BinaryMessage)
}

// WriteText sends text down side channel
func (conn *WSConn) WriteText(b []byte) (n int, err error) {
	return conn.write(b, websocket.TextMessage)
}

func (conn *WSConn) WritePing(b []byte) (err error) {
	_, err = conn.write(b, websocket.PingMessage)
	return
}

func (conn *WSConn) WritePong(b []byte) (err error) {
	_, err = conn.write(b, websocket.PongMessage)
	return
}

func (conn *WSConn) WriteClose(b []byte) (err error) {
	_, err = conn.write(b, websocket.CloseMessage)
	return
}

// Central write function to handle mutexes and stuff
// We have ot have mutex in package because some package that grabs
// our wsconn like a net.conn definitely won't respect our side channels (text, ping, etc)
func (conn *WSConn) write(b []byte, mt int) (n int, err error) {
	conn.writeMutex.Lock()
	defer conn.writeMutex.Unlock()

	if mt == websocket.PingMessage || mt == websocket.PongMessage || mt == websocket.CloseMessage {
		err = conn.Conn.WriteMessage(mt, b)
		return 0, err
	}
	wc, err := conn.Conn.NextWriter(mt)
	if err != nil {
		return 0, err
	}
	n, err = wc.Write(b)
	if err != nil {
		return n, err
	}
	err = wc.Close() // close the writer, open a new one the next write.
	// this seems inefficient, but we don't know what kind of message we're going to send next
	return n, err
}

// Close is a simple wrap for the underlying websocket.Conn
func (conn *WSConn) Close() error {
	conn.TextBuffer = nil // probalby not necessary
	return conn.Conn.Close()
}

// LocalAddr is a simple wrap for the underlying websocket.Conn
func (conn *WSConn) LocalAddr() net.Addr {
	return conn.Conn.LocalAddr()
}

// RemoteAddr is a simple wrap for the underlying websocket.Conn
func (conn *WSConn) RemoteAddr() net.Addr {
	return conn.Conn.RemoteAddr()
}

// SetDeadline is a simple wrap for the underlying websocket.Conn
func (conn *WSConn) SetDeadline(t time.Time) error {
	if err := conn.SetReadDeadline(t); err != nil {
		defaultLogger.Error("SetReadDeadline(): " + err.Error())
		return err
	}
	if err := conn.SetWriteDeadline(t); err != nil {
		defaultLogger.Error("SetWriteDeadline(): " + err.Error())
		return err
	}
	return nil
}

// SetReadDeadline is a simple wrap for the underlying websocket.Conn
func (conn *WSConn) SetReadDeadline(t time.Time) error {
	if err := conn.SetReadDeadline(t); err != nil {
		defaultLogger.Error("SetReadDeadline(): " + err.Error())
		return err
	}
	return nil
}

// SetWriteDeadline is a simple wrap for the underlying websocket.Conn
func (conn *WSConn) SetWriteDeadline(t time.Time) error {
	if err := conn.SetWriteDeadline(t); err != nil {
		defaultLogger.Error("SetWriteDeadline():" + err.Error())
		return err
	}
	return nil
}

// USEFUL SERVER FUNCTIONS
type Upgrader struct {
	websocket.Upgrader
}

var upgrader websocket.Upgrader

func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WSConn, error) {
	conn, err := u.Upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return New(conn)
}

// THIS IS FOR THE CLIENT
func Dial(urlStr string, requestHeader http.Header) (*WSConn, *http.Response, error) {
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(urlStr, requestHeader)
	if err != nil {
		defaultLogger.Error("dialer.Dial: " + err.Error())
		return nil, nil, err
	}
	wsconn, err := New(conn)
	return wsconn, resp, err // Not really sure about resp here- why does Dail* return it?
}

// DialContext wraps websockets.DialContext
func DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*WSConn, *http.Response, error) {
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.DialContext(ctx, urlStr, requestHeader)
	if err != nil {
		defaultLogger.Error("dialer.DialContext: " + err.Error())
		return nil, nil, err
	}
	wsconn, err := New(conn)
	return wsconn, resp, err // Not really sure about resp here- why does Dail* return it?

}
