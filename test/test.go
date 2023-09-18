package unit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
)

var (
	// router
	router http.Handler

	// customed request headers for token authorization and so on
	myHeaders = make(map[string]string, 0)

	logging *log.Logger
)

// set the router
func SetRouter(r http.Handler) {
	router = r
}

// set the log
func SetLog(l *log.Logger) {
	logging = l
}

// add custom request header
func AddHeader(key, value string) {
	myHeaders[key] = value
}

// printf log
func printfLog(format string, v ...interface{}) {
	if logging == nil {
		return
	}

}

// invoke handler
func invokeHandler(req *http.Request) (bodyByte []byte, err error) {

	// initialize response record
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// extract the response from the response record
	result := w.Result()
	defer result.Body.Close()

	// extract response body
	bodyByte, err = ioutil.ReadAll(result.Body)
	return
}

func TestHandler(method string, url string, parameter interface{}) (bodyByte []byte, err error) {
	// check whether the router is nil
	if router == nil {
		err = errors.New("router not set")
		return
	}

	var (
		contentBuffer *bytes.Buffer
		jsonBytes     []byte
		request       *http.Request
	)
	jsonBytes, err = json.Marshal(parameter)
	if err != nil {
		return
	}
	contentBuffer = bytes.NewBuffer(jsonBytes)
	request, err = http.NewRequest(string(method), url, contentBuffer)
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=utf-8")

	bodyByte, err = invokeHandler(request)
	printfLog("method: %v, url: %v, parameter: , response: %v", method, url, string(bodyByte))
	return
}
