package main

import (
	"context"
	"flag"
	"log"

	"github.com/mangodx/AccessTunnel/sshoverws"
	"github.com/gorilla/mux"
)

var hostPrivateKey = flag.String("hostkey", "", "Path to the your private key")

func init(){
	flag.Parse()
	if (len(*hostPrivateKey) == 0) {
		log.Fatalf("You must set a hostkey with -hostkey")
	}
}

	// ESTABLISH SSH CONNETION DOWNPIPE

func handleProxy(w http.ResponseWriter, r *http.Request) {
	conn, err := sshoverws.Upgrade(w, r)
	if err != nil {
		log.Printf("Error on upgrade (%s)", err)
	}
	// read, write, read, write, read, write, read (if close) close
}

func main(){

}

