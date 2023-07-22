package wsconn

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
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
	defaultLogger.Info("SetDefaultLogger: Default Logger Set")
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
	TextChan   chan int

	// WRITING STUFF
	writeMutex sync.Mutex

	//
	allowClose atomic.Bool
	closeMutex sync.Mutex
}

// New returns an initialized *WSConn
func New(conn *websocket.Conn) (wsconn *WSConn, err error) {
	wsconn = &WSConn{Conn: conn, r: nil, mt: 0, TextBuffer: new(bytes.Buffer)}
	wsconn.Conn.SetPingHandler(func(message string) error {
		defaultLogger.Info("Ping In: " + message)
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

// Read() satisfying net.Conn
func (conn *WSConn) Read(b []byte) (n int, err error) {
	for {
		if conn.mt == 0 { // Need NextReader()
			mt, r, err := conn.Conn.NextReader() // Any error NextReader receives is fatal
			if err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseAbnormalClosure,
				) {
					defaultLogger.Error("WSConn.Read() Received A NextReader Close Error: " + err.Error())
					return 0, err
				}
				return 0, err
				defaultLogger.Error("WSConn.Read() Received A NextReader Non-Close Error: " + err.Error())
			}

			conn.mt = mt
			conn.r = r

			if conn.mt > websocket.BinaryMessage {
				return 0, errors.New("wsconn.Read(): Wrong error type received")
			}
		}
		for {
			n, err = conn.r.Read(b)
			defaultLogger.Info("Read: " + strconv.Itoa(n))
			if err != nil || b == nil {
				conn.r = nil
				mt = conn.mt
				conn.mt = 0
				if err == io.EOF {
					defaultLogger.Error("WSConn.Read() EOF End Of Frame")
					if n != 0 {
						if mt == websocket.BinaryMessage {
							return n, nil
						} // else text message
						_, err := conn.TextBuffer.Write(b[0:n]) // Length TODO
						if err != nil {
							return n, err // TODO not technically correct
						}
						conn.TextChan <- n
					}
					break
				} else {
					defaultLogger.Error("WSConn.Read() error: " + err.Error())
					return n, err
				}
			}
			if conn.mt == websocket.TextMessage {
				if conn.TextChan != nil {
					_, err := conn.TextBuffer.Write(b[0:n]) // Length TODO
					if err != nil {
						return n, err // TODO not technically correct
					}
					conn.TextChan <- n
				} else {
					// return 0, errors.New("wsconn.TextChan: Doesn't exist")
					// Just keep looking for binary messages
					conn.TextBuffer.Reset()
				}
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
	defaultLogger.Info("Ping Out: " + string(b[:]))
	_, err = conn.write(b, websocket.PingMessage)
	return
}

func (conn *WSConn) WritePong(b []byte) (err error) {
	defaultLogger.Info("Pong Out: " + string(b[:]))
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

func (conn *WSConn) AllowClose() {
	conn.allowClose.Store(true)
}

// Close is a simple wrap for the underlying websocket.Conn
func (conn *WSConn) Close() error {
	if !conn.allowClose.Load() {
		return nil
	}
	conn.closeMutex.Lock()
	defer conn.closeMutex.Unlock()
	defaultLogger.Info("Closing underlying connection")
	err := conn.Conn.Close()
	if conn.TextChan != nil {
		defaultLogger.Info("Closing channel")
		close(conn.TextChan)
		conn.TextChan = nil
	}
	defaultLogger.Info("Freeing TextBuffer Pointer")
	conn.TextBuffer = nil // probalby not necessary
	return err
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
