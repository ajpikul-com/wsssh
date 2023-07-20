package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ajpikul-com/ilog"
	"github.com/ajpikul-com/wsssh/wsconn"
)

var defaultLogger ilog.LoggerInterface

func init() {
	defaultLogger = &ilog.ZapWrap{Sugar: true}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	wsconn.SetDefaultLogger(defaultLogger)
}

func dumpResponse(resp *http.Response) {
	if resp != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			defaultLogger.Error(err.Error())
			// return ?
		}
		extra := "Body:\n" + string(b)
		defaultLogger.Error("HTTP Response Info: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status + "\n" + extra)
	}
}
