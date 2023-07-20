package wsconn

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
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

// SetDefaultLogger attaches a logger to the lib. See github.com/ayajyt/ilog
func SetDefaultLogger(newLogger ilog.LoggerInterface) {
	defaultLogger = newLogger
	defaultLogger.Info("Default Logger Set")
}

// WSTransport satisfied the net.Conn interface by wrapping the websocket.Conn
type WSTransport struct {
	Conn *websocket.Conn
	r    io.Reader // it has to save the reader because websocket.Conn forces you to reuse readers
	mt   int
}

// Read wraps websockets read so that the whole connection is treated as a continue stream, throwing out any EOFs.  So if there is a legit EOF, it won't work- it should be a Close handshake anyway.
func (wst *WSTransport) Read(b []byte) (n int, err error) {
	defaultLogger.Info("In WSTransport.Read")
	if wst.r == nil {
		defaultLogger.Info("WSTransport Reader was nil, calling NextReader()")
		var mt int
		mt, r, err := wst.Conn.NextReader() // Errors from here are fatal, connection must be reset
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseAbnormalClosure,
			) {
				defaultLogger.Info("WSTransport.NextReader returned close, so websockets over.")
				return 0, err
			}
			defaultLogger.Error("WSTransport.Read.NextReader(): " + err.Error())
			return 0, err
		}
		wst.mt = mt
		wst.r = r
		var mtStr string
		if mt == websocket.TextMessage {
			mtStr = "TextMessage"
		} else if mt == websocket.BinaryMessage {
			mtStr = "BinaryMessage"
		} else if mt == websocket.CloseMessage {
			mtStr = "CloseMessage"
		} else if mt == websocket.PingMessage {
			mtStr = "PingMessage"
		} else if mt == websocket.PongMessage {
			mtStr = "PongMessage"
		}
		defaultLogger.Info("MessageType from websockets: " + mtStr)
		if wst.mt > websocket.BinaryMessage {
			defaultLogger.Error("WSTransport.Read() received a non-binary/text message: " + mtStr)
			return 0, errors.New("Wrong error type received")
		}
	}
	if wst.mt == websocket.TextMessage {
		n, err = wst.r.Read(b)
		// TODO: circular buffer with side messages
		if err != nil {
			defaultLogger.Error("websocket.NextReader's io.Reader.Read():" + err.Error())
		}
	} else {
		n, err = wst.r.Read(b)
	}
	if b != nil {
		defaultLogger.Info("Packet Received: (amount=" + strconv.Itoa(n) + ")\n" + strconv.Quote(string(bytes.Trim(b, "\x00"))))
	}
	if err != nil {
		if err == io.EOF {
			err = nil
		} else {
			defaultLogger.Error("WSTransport.Read(): " + err.Error())
		}
		wst.r = nil
		wst.mt = 0
	}
	if wst.mt == websocket.TextMessage {
		return 0, nil
	}
	return n, err
}

// Write does a write but facilitates getting a reader
// This needs to be single threaded right?
func (wst *WSTransport) Write(b []byte) (n int, err error) {
	defaultLogger.Info("WSTransport.Write() called")
	wc, err := wst.Conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	n, err = wc.Write(b)
	if err != nil {
		return n, err
	}
	err = wc.Close()
	return n, err
}

func (wst *WSTransport) WriteText(s string) error {
	defaultLogger.Info("Sending text message via websocket: " + s)
	var err error = nil
	if err = wst.Conn.WriteMessage(websocket.TextMessage, []byte(s)); err != nil {
		defaultLogger.Error("WSTransport.WriteText(): " + err.Error())
	}
	return err
}

// Close is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) Close() error {
	defaultLogger.Info("Calling sshoverws.WSTransport.Close()")
	return wst.Conn.Close()
}

// LocalAddr is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) LocalAddr() net.Addr {
	addr := wst.Conn.LocalAddr()
	defaultLogger.Info("Calling sshoverws.WSTransport.LocalAddr()-> " + addr.Network() + ": " + addr.String())
	return addr
}

// RemoteAddr is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) RemoteAddr() net.Addr {
	addr := wst.Conn.RemoteAddr()
	defaultLogger.Info("Calling sshoverws.WSTransport.RemoteAddr()-> " + addr.Network() + ": " + addr.String())
	return addr
}

// SetDeadline is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) SetDeadline(t time.Time) error {
	defaultLogger.Info("Calling WSTransport.SetDeadline()")
	if err := wst.SetReadDeadline(t); err != nil {
		defaultLogger.Error("SetReadDeadline(): " + err.Error())
		return err
	}
	if err := wst.SetWriteDeadline(t); err != nil {
		defaultLogger.Error("SetWriteDeadline(): " + err.Error())
		return err
	}
	return nil
}

// SetReadDeadline is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) SetReadDeadline(t time.Time) error {
	defaultLogger.Info("Calling WSTransport.SetReadDeadline()")
	if err := wst.SetReadDeadline(t); err != nil {
		defaultLogger.Error("SetReadDeadline(): " + err.Error())
		return err
	}
	return nil
}

// SetWriteDeadline is a simple wrap for the underlying websocket.Conn
func (wst *WSTransport) SetWriteDeadline(t time.Time) error {
	defaultLogger.Info("Calling WSTransport.SetWriteDeadline()")
	if err := wst.SetWriteDeadline(t); err != nil {
		defaultLogger.Error("SetWriteDeadline():" + err.Error())
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
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true // so this basically allows requests from any origin? okay...
		},
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*WSTransport, error) {
	defaultLogger.Info("Calling WSTransport.Upgrade(w,r)")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		defaultLogger.Error("Upgrade: " + err.Error())
	}
	return WrapConn(conn), err
}

// Dial wraps websockets.dial
func Dial(urlStr string, requestHeader http.Header) (*WSTransport, *http.Response, error) {
	defaultLogger.Info("Calling WSTransport.Dial")
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(urlStr, requestHeader)
	if err != nil {
		defaultLogger.Error("dialer.Dial: " + err.Error())
		return nil, nil, err
	}
	return WrapConn(conn), resp, err // Not really sure about resp here- why does Dail* return it?
}

// DialContext wraps websockets.DialContext
func DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*WSTransport, *http.Response, error) {
	defaultLogger.Info("Calling WSTransport.DialContext")
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.DialContext(ctx, urlStr, requestHeader)
	if err != nil {
		defaultLogger.Error("dialer.DialContext: " + err.Error())
		return nil, nil, err
	}
	return WrapConn(conn), resp, err // Not really sure about resp here- why does Dail* return it?
}

// WrapConn takes a websockets connection and makes it a proper stream. Helper functions are provided (dial, upgrade)
func WrapConn(conn *websocket.Conn) *WSTransport {
	// I guess we get net.Addr from here
	defaultLogger.Info("In WSTransport.WrapConn")
	return &WSTransport{Conn: conn, r: nil}
}