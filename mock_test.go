package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/url"
)

func b64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

type MockAuthenticatorInvocation struct {
	user        string
	pass        string
	extraGroups []string
}

type MockAuthenticator struct {
	calls       []MockAuthenticatorInvocation
	returnValue int
}

func (ta *MockAuthenticator) Authenticate(user, pass string, extraGroups []string) int {
	ta.calls = append(ta.calls, MockAuthenticatorInvocation{
		user:        user,
		pass:        pass,
		extraGroups: extraGroups,
	})
	return ta.returnValue
}

func testAuther(returnValue int) *MockAuthenticator {
	return &MockAuthenticator{
		calls:       make([]MockAuthenticatorInvocation, 0),
		returnValue: returnValue,
	}
}

type TestResponseWriter struct {
	headers    http.Header
	statusCode int
	buffer     *bytes.Buffer
}

func testRequest(method string, uri string, user string, pass string) *http.Request {
	reqUrl, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	req := &http.Request{
		Method:     method,
		URL:        reqUrl,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if user != "" && pass != "" {
		req.Header.Add("Authorization", "Basic "+b64(user+":"+pass))
	}

	return req
}

func testResponse() *TestResponseWriter {
	response := &TestResponseWriter{
		headers:    make(http.Header),
		statusCode: -1,
		buffer:     bytes.NewBuffer(nil),
	}
	return response
}

func (trw *TestResponseWriter) Header() http.Header {
	return trw.headers
}

func (trw *TestResponseWriter) Write(data []byte) (int, error) {
	return trw.buffer.Write(data)
}

func (trw *TestResponseWriter) WriteHeader(statusCode int) {
	trw.statusCode = statusCode
}
