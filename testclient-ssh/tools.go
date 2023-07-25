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
	defaultLogger = &ilog.SimpleLogger{}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	wsconn.SetDefaultLogger(defaultLogger)
	// TODO: test non debug
	// TODO: test sugarred logger for function
}

func dumpResponse(resp *http.Response) string {
	if resp != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			defaultLogger.Error(err.Error())
			// return ?
		}
		extra := "Body:\n" + string(b)
		return "HTTP Response Info: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status + "\n" + extra
	}
	return ""
}
