package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ajpikul-com/wsssh/wsconn"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var returnServerConn func(chan error, chan bool) = func(errchan chan error, clientchan chan bool) http.Handler {

	return func(w http.ResponseWriter, r *http.Request) {
		clientchan <- true
		upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true
		},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errchan <- err
			return
		}
		wsconn, err := wsconn.New(conn)
		if err != nil {
			errchan <- err
			return
		}

		var n int = 0
		readBuffer := make([]byte, 1024)
		for err == nil {
			for n, err = wsconn.Read(readBuffer); n != 0; n, err = wsconn.Read(readBuffer) {
				if err != nil {
					errchan <- err
					break
				}
			}
		}

		defer func() {
			if err := wsconn.Close(); err != nil {
				errchan <- err
				return
			}
		}()
	}
}

func serve(hostport string, errchan chan error, clientchan chan bool) {
	m := mux.NewRouter()
	m.HandleFunc("/", ServeWSConn)
	s := &http.Server{
		Addr:           string,
		Handler:        m,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		errchan <- err
		return
	}
	return // server is done
}
