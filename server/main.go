package main

import (
	"time"
	"net/http"
	"flag"
	"log"

	"github.com/mangodx/AccessTunnel/sshoverws"
	"github.com/gorilla/mux"
)

var hostPrivateKey = flag.String("hostkey", "", "Path to the your private key")

func init(){
	flag.Parse()
	if (len(*hostPrivateKey) == 0) {
		//log.Fatalf("You must set a hostkey with -hostkey")
	}
}

	// ESTABLISH SSH CONNETION DOWNPIPE

func handleProxy(w http.ResponseWriter, r *http.Request) {
	log.Printf("Req: %s %s", r.Host, r.URL.Path)
	conn, err := sshoverws.Upgrade(w, r)
	if err != nil {
		log.Printf("Error on upgrade (%s)", err)
	}
	conn.Write([]byte("I did it!"))
	conn.Write([]byte("I did it!"))
	conn.Write([]byte("I did it!"))
	conn.Write([]byte("I did it!"))
	conn.Write([]byte("I did it!"))
	buffer := make([]byte, 32*256)
	conn.Read(buffer)
	conn.Write([]byte("Got it"))
// read, write, read, write, read, write, read (if close) close
}

func main(){
	m := mux.NewRouter()
	m.HandleFunc("/", handleProxy)
	s := &http.Server {
		Addr: "127.0.0.1:2223",
		Handler:	m,
		ReadTimeout:	10 * time.Second,
		WriteTimeout:	10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}

