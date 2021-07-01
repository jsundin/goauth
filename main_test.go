package main

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}

func TestAuthenticatorError(t *testing.T) {
	req := testRequest("GET", "http://testsite.com/auth?groups=g1,g2", "username", "passwd")
	res := testResponse()
	auther := testAuther(AuthError)

	authenticationHandler(auther, "realm", res, req)

	if len(auther.calls) != 1 {
		t.Errorf("expected 1 call to auther, but got %d", len(auther.calls))
	}

	assert.Equal(t, 1, len(auther.calls), "expected one call to auther")
	assert.Equal(t, 2, len(auther.calls[0].extraGroups), "expected two extragroups")
	assert.Equal(t, "g1", auther.calls[0].extraGroups[0])
	assert.Equal(t, "g2", auther.calls[0].extraGroups[1])
	assert.Equal(t, "username", auther.calls[0].user)
	assert.Equal(t, "passwd", auther.calls[0].pass)

	assert.Equal(t, 500, res.statusCode, "bad status code")
	assert.Equal(t, "Internal server error", res.buffer.String(), "bad response body")
}

func TestBadGroupsParameter(t *testing.T) {
	req := testRequest("GET", "http://testsite.com/auth?groups=a&groups=b", "", "")
	res := testResponse()
	auther := testAuther(AuthError)

	authenticationHandler(auther, "realm", res, req)

	assert.Equal(t, 0, len(auther.calls), "expected no calls to auther")
	assert.Equal(t, 400, res.statusCode, "bad status code")
	assert.Equal(t, "Missing 'groups' parameter", res.buffer.String())
}

func TestMissingAuth(t *testing.T) {
	req := testRequest("GET", "http://testsite.com/auth", "", "")
	res := testResponse()
	auther := testAuther(AuthError)

	authenticationHandler(auther, "test-realm", res, req)

	assert.Equal(t, 0, len(auther.calls), "expected no calls to auther")
	assert.Equal(t, 401, res.statusCode, "bad status code")
	assert.Equal(t, "Unauthorized", res.buffer.String(), "bad response body")
	assert.Equal(t, 1, len(res.headers), "expected one header")
	assert.Equal(t, "Basic realm=\"test-realm\"", res.headers.Get("WWW-Authenticate"), "bad authentication header in response")
}

func TestUnsuccessfulAuth(t *testing.T) {
	req := testRequest("GET", "http://testsite.com/auth?groups=addgrp", "testuser", "testpasswd")
	res := testResponse()
	auther := testAuther(AuthFailed)

	authenticationHandler(auther, "test-realm", res, req)

	assert.Equal(t, 1, len(auther.calls), "expected one call to auther")
	assert.Equal(t, 1, len(auther.calls[0].extraGroups), "expected one extra group")
	assert.Equal(t, "addgrp", auther.calls[0].extraGroups[0], "expected one extra group")
	assert.Equal(t, "testuser", auther.calls[0].user, "bad user")
	assert.Equal(t, "testpasswd", auther.calls[0].pass, "bad password")

	assert.Equal(t, 401, res.statusCode, "bad status code")
	assert.Equal(t, "Unauthorized", res.buffer.String(), "bad response body")
	assert.Equal(t, 1, len(res.headers), "expected one header")
	assert.Equal(t, "Basic realm=\"test-realm\"", res.headers.Get("WWW-Authenticate"), "bad authentication header in response")
}

func TestSuccessfulAuth(t *testing.T) {
	req := testRequest("GET", "http://testsite.com/auth?groups=addgrp", "testuser", "testpasswd")
	res := testResponse()
	auther := testAuther(AuthSuccess)

	authenticationHandler(auther, "test-realm", res, req)

	assert.Equal(t, 1, len(auther.calls), "expected one call to auther")
	assert.Equal(t, 1, len(auther.calls[0].extraGroups), "expected one extra group")
	assert.Equal(t, "addgrp", auther.calls[0].extraGroups[0], "expected one extra group")
	assert.Equal(t, "testuser", auther.calls[0].user, "bad user")
	assert.Equal(t, "testpasswd", auther.calls[0].pass, "bad password")

	assert.Equal(t, -1, res.statusCode, "bad status code (should not be set)")
	assert.Equal(t, "", res.buffer.String())
	assert.Equal(t, 0, len(res.headers), "expected no headers")
}
