package main

import (
	"net/http"
	// "sync"
	//"strconv"
	"time"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var defaultLogger ilog.LoggerInterface
var HostPort string = "127.0.0.1:2223"

func init() {
	defaultLogger = &ilog.SimpleLogger{}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	wsconn.SetDefaultLogger(defaultLogger)
}

func WriteText(conn *wsconn.WSConn) {
	for {
		_, err := conn.WriteText([]byte("Test Message")) // TODO Can we be sure this will write everything
		if err != nil {
			defaultLogger.Error("WriteText: wsconn.WriteText(): " + err.Error())
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func WriteBinary(conn *wsconn.WSConn) {
	for i := 0; i < 3; i++ {
		_, err := conn.Write([]byte("12345678")) // TODO sure it will write everyting?
		if err != nil {
			defaultLogger.Error("WriteBinary: wsconn.Write(): " + err.Error())
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func ReadTexts(conn *wsconn.WSConn) {
	/*
		textChan := make(chan int)
		conn.TextChan = textChan
		p := make([]byte, 1024)
		for _ = range textChan {
			for conn.TextBuffer.Len() > 0 {
				n, err := conn.TextBuffer.Read(p)
				defaultLogger.Info("ReadTexts: " + string(p[0:n]))
				defaultLogger.Info("ReadTexts Remaining: " + strconv.Itoa(conn.TextBuffer.Len()))
				if err != nil {
					defaultLogger.Error("ReadTexts: TextBuffer.Read():" + err.Error())
					break
				}
			}
		}
		defaultLogger.Info("ReadTexts Channel Closed")
		// The channel has been closed by someone else
	*/
	// WE DON'T DO IT LIKE THIS ANYMORE, SSH CLIENT/SERVER HAS IT RIGHT
}

func ReadBinary(conn *wsconn.WSConn) {
	var n int = 0
	readBuffer := make([]byte, 256)
	var err error = nil
	for err == nil {
		for n, err = conn.Read(readBuffer); n != 0; n, err = conn.Read(readBuffer) {
			defaultLogger.Info("ReadBinary: " + string(readBuffer[0:n]))
			if err != nil {
				// Errors usually won't get here because n = 0
				defaultLogger.Error("ReadBinary: wsconn.Read():" + err.Error())
				break
			}
		}
		// The error we got was fatal
		// It was probably a close close error, and that's fine
		// Either way, everythings fucked and we wait for reconnect
	}
	defaultLogger.Error("ReadBinary closing with err: " + err.Error())
}

func ServeWSConn(w http.ResponseWriter, r *http.Request) {
	defaultLogger.Info("Server: Incoming Req: " + r.Host + ", " + r.URL.Path)
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		defaultLogger.Error("Server: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	wsconn, err := wsconn.New(conn)
	if err != nil {
		panic(err.Error())
	}

	go ReadTexts(wsconn)
	go WriteBinary(wsconn)
	ReadBinary(wsconn) // If it's a go routine, we have to wait for it // If it's a go routine, we have to wait for it

	defer func() {
		defaultLogger.Info("Server: Closing WSConn")
		if err := wsconn.Close(); err != nil {
			defaultLogger.Error("Server: " + err.Error())
		}
	}()
}

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/", ServeWSConn)
	s := &http.Server{
		Addr:           HostPort,
		Handler:        m,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	defaultLogger.Info("Server: Initiating server")
	err := s.ListenAndServe()
	if err != nil {
		defaultLogger.Error("Server: http.Server.ListenAndServe: " + err.Error())
	}
	defaultLogger.Info("Server exiting")
}
