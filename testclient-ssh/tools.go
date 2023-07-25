package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ajpikul-com/ilog"
	//"github.com/ajpikul-com/wsssh/wsconn"
)

var defaultLogger ilog.LoggerInterface

func init() {
	defaultLogger = &ilog.ZapWrap{}
	err := defaultLogger.Init()
	if err != nil {
		panic(err)
	}
	/*packageLogger := &ilog.SimpleLogger{}
	packageLogger.Level(ilog.INFO)
	packageLogger.Init()
	wsconn.SetDefaultLogger(packageLogger)*/
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
