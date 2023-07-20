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
	defaultLogger.Info("wsconn.Read()")
	if wst.r == nil {
		defaultLogger.Info("wsconn.Read(): Need new reader, calling NextReader()")
		var mt int
		mt, r, err := wst.Conn.NextReader() // Errors from here are fatal, connection must be reset
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseAbnormalClosure,
			) {
				defaultLogger.Info("wsconn.Read().NextReader(): A Close Return: " + err.Err())
				return 0, err
			}
			defaultLogger.Error("wsconn.Read().NextReader(): " + err.Error())
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
		defaultLogger.Info("wsconn.Read(): MessageType: " + mtStr)
		if wst.mt > websocket.BinaryMessage {
			defaultLogger.Error("wsconn.Read(): unexpected message type")
			// is it true we never receive control from NextReader
			return 0, errors.New("wsconn.Read(): Wrong error type received")
		}
	}
	n, err = wst.r.Read(b)
	if b != nil {
		defaultLogger.Info("wsconn.Read(): Packet Received: (amount=" + strconv.Itoa(n) + ")\n" + strconv.Quote(string(bytes.Trim(b, "\x00"))))
	}
	if err != nil {
		wst.r = nil // set wst.mt to 0 later after processing
		if err == io.EOF {
			defaultLogger.Info("wsconn.Read(): reached EOF")
			err = nil
			// This is not a real error tha twe want to report from read, it's normal.
			// We give what we can give, you call again
		} else {
			defaultLogger.Error("wsconn.Read(): NextReader's io.Reader.Read() " + err.Error())
		}
	}
	if wst.mt == websocket.TextMessage {
		// TODO: don't report this directly
		// Maybe we need seperate ReadText()
		// Or just dump this all into hooks?
		return 0, nil
	}

	if wst.r == nil {
		wst.mt = 0
	}
	// n is better than err for detecting if read is done
	// but how should the caller loop on this, not sure yet
	return n, err
}

// Are there race conditions here
func (wst *WSTransport) Write(b []byte) (n int, err error) {
	defaultLogger.Info("wsconn.Write(): WSTransport.Write() called")
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
	defaultLogger.Info("wsconn.WriteText(): Sending text message via websocket: " + s)
	var err error = nil
	if err = wst.Conn.WriteMessage(websocket.TextMessage, []byte(s)); err != nil {
		defaultLogger.Error("wsconn.WriteText(): WSTransport.WriteText(): " + err.Error())
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
