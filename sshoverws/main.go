package sshoverws

import (
	"net"
	"time"
	"errors"
	"net/http"
	"context"
	"strconv"
	"io"

	"github.com/ayjayt/ilog"

	"github.com/gorilla/websocket"
)

var defaultLogger ilog.LoggerInterface

func init() {
	if defaultLogger == nil {
		defaultLogger = new(ilog.EmptyLogger)
	}
}

// SetDefaultLogger attaches a logger to the lib. See github.com/ayajyt/ilog
func SetDefaultLogger(newLogger ilog.LoggerInterface) {
	defaultLogger = newLogger
	defaultLogger.Info("INFO: Default Logger Set")
}

// WSTransport satisfied the net.Conn interface by wrapping the websocket.Conn
type WSTransport struct {
	conn *websocket.Conn
	r io.Reader // it has to save the reader because websocket.Conn forces you to reuse readers
}

// Read wraps websockets read so that the whole connection is treated as a continue stream, throwing out any EOFs.  So if there is a legit EOF, it won't work- it should be a Close handshake anyway. 
func (wst *WSTransport) Read(b []byte) (n int, err error) {
	defaultLogger.Info("INFO: WSTransport.Read")
	if wst.r == nil {
		defaultLogger.Info("INFO: reader was nil, calling next")
		var mt int
		mt, r, err := wst.conn.NextReader() // Errors from here are fatal, connection must be reset
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseAbnormalClosure,
			) {
				defaultLogger.Info("INFO: WSTransport.NextReader returned close.")
				return 0, err // TODO: probably want to translate close 
			}
			defaultLogger.Error("AccessTunnel/sshoverws/main.go WSTransport.Read.NextReader: " + err.Error())
			return 0, err // What other errors are we dealing with?
		}
		if mt != websocket.BinaryMessage {
			var mtStr string
			if mt == 1 {
				mtStr = "TextMessage"
			} else if mt == 2 {
				mtStr = "BinaryMessage"
			} else if mt == 8 {
				mtStr = "CloseMessage"
			} else if mt == 9 {
			  mtStr = "PingMessage"
			} else if mt == 10 {
				mtStr = "PongMessage"
			}
			defaultLogger.Error("AccessTunnel/sshoverws/main.go WSTransport.Read() received a non-binary message: " + mtStr)
			return 0, errors.New("Wrong error type received")
		}
		wst.r = r
	}
	defaultLogger.Info("INFO: Reading")
	n, err = wst.r.Read(b) // Read errors are not stated to be fatal but except for EOF seem pretty screwed
	defaultLogger.Info("INFO >Read: " + strconv.Itoa(n) + ", " + err.Error())
	if err != nil {
		if err == io.EOF { // Not sure what else it could be, check to see if fatal
			err = nil
		}
		wst.r = nil
		defaultLogger.Error("AccessTunnel/sshoverws/main.go WSTransport.Read() received error on read: " + err.Error())
	}
	return n, err
}

// Write does a write but facilitates getting a reader
// This needs to be single threaded right?
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

// Close is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) Close() error {
	defaultLogger.Info("INFO: Calling sshoverws.WSTransport.Close()")
	return wst.conn.Close()
}

// LocalAddr is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) LocalAddr() net.Addr {
	addr := wst.conn.LocalAddr()
	defaultLogger.Info("INFO: Calling sshoverws.WSTransport.LocalAddr()-> " + addr.Network() + ": " + addr.String())
	return addr
}

// RemoteAddr is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) RemoteAddr() net.Addr {
	addr := wst.conn.RemoteAddr()
	defaultLogger.Info("INFO: Calling sshoverws.WSTransport.RemoteAddr()-> " + addr.Network() + ": " + addr.String())
	return addr
}

// SetDeadline is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) SetDeadline(t time.Time) error {
	defaultLogger.Info("INFO: Calling SetDeadline()")
	if err := wst.SetReadDeadline(t); err != nil {
		defaultLogger.Error("AccessTunnel/sshoverws/main.go SetReadDeadline " + err.Error())
		return err
	}
	if err := wst.SetWriteDeadline(t); err != nil {
		defaultLogger.Error("AccessTunnel/sshoverws/main.go SetWriteDeadline " + err.Error())
		return err
	}
	return nil
}

// SetReadDeadline is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) SetReadDeadline(t time.Time) error {
	defaultLogger.Info("INFO: Calling SetReadDeadline()")
	if err := wst.SetReadDeadline(t); err != nil {
		defaultLogger.Error("AccessTunnel/sshoverws/main.go SetReadDeadline " + err.Error())
		return err
	}
	return nil
}

// SetWriteDeadline is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) SetWriteDeadline(t time.Time) error {
	defaultLogger.Info("INFO: Calling SetWriteDeadline()")
	if err := wst.SetWriteDeadline(t); err != nil {
		defaultLogger.Error("AccessTunnel/sshoverws/main.go SetWriteDeadline " + err.Error())
		return err
	}
	return nil
}

// Probably not the best way to provide this, but I can't figure out why exactly besides aesthetic. It does make the API easier.
// Question is: will client need access to upgrader? Don't want to try and replace 
var upgrader websocket.Upgrader

func init() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:   1024, // is it reasonable size? how do we tune
		WriteBufferSize:  1024,
		HandshakeTimeout: 10*time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true // so this basically allows requests from any origin? okay...
		},
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*WSTransport, error) {
	defaultLogger.Info("INFO: Calling Upgrade(w,r)")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		defaultLogger.Error("AccessTunnel/sshoverws/main.go Upgrade err: " + err.Error())
	}
	return WrapConn(conn), err
}

// Dial wraps websockets.dial
func Dial(urlStr string, requestHeader http.Header) (*WSTransport, *http.Response, error) {
	defaultLogger.Info("INFO: Calling Dial")
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(urlStr, requestHeader)
	if err != nil {
		defaultLogger.Error("AccessTunnel/sshoverws.Dial(): dialer.Dial error: " + err.Error())
		return nil, nil, err
	}
	return WrapConn(conn), resp, err // Not really sure about resp here- why does Dail* return it?
}

// DialContext wraps websockets.DialContext
func DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*WSTransport, *http.Response, error) {
	defaultLogger.Info("INFO: Calling DialContext")
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.DialContext(ctx, urlStr, requestHeader)
	if err != nil {
		defaultLogger.Error("iAccessTunnel/sshoverws.DialContext(): dialer.DialContext error: " + err.Error())
		return nil, nil, err
	}
	return WrapConn(conn), resp, err // Not really sure about resp here- why does Dail* return it?
}

// WrapConn takes a websockets connection and makes it a proper stream. Helper functions are provided (dial, upgrade)
func WrapConn(conn *websocket.Conn) *WSTransport {
	// I guess we get net.Addr from here
	defaultLogger.Info("INFO: In WrapConn")
	return &WSTransport{conn:conn, r:nil}
}


