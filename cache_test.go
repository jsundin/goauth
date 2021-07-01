package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	auther := testAuther(AuthSuccess)
	ca := NewCachedAuthenticator(auther, 2*time.Second)

	result1 := ca.Authenticate("user1", "pass", []string{"g1"})
	result2 := ca.Authenticate("user1", "pass", []string{"g1"}) // should not be checked
	result3 := ca.Authenticate("user2", "pass", []string{"g1"})
	result4 := ca.Authenticate("user2", "pass", []string{"g2"})

	assert.Equal(t, 3, len(auther.calls), "expected 2 calls to mock auther")
	assert.Equal(t, "user1", auther.calls[0].user)
	assert.Equal(t, []string{"g1"}, auther.calls[0].extraGroups)
	assert.Equal(t, "user2", auther.calls[1].user)
	assert.Equal(t, []string{"g1"}, auther.calls[1].extraGroups)
	assert.Equal(t, "user2", auther.calls[2].user)
	assert.Equal(t, []string{"g2"}, auther.calls[2].extraGroups)

	assert.Equal(t, AuthSuccess, result1)
	assert.Equal(t, AuthSuccess, result2)
	assert.Equal(t, AuthSuccess, result3)
	assert.Equal(t, AuthSuccess, result4)
}

func TestCacheExpiration(t *testing.T) {
	auther := testAuther(AuthSuccess)
	ca := NewCachedAuthenticator(auther, 500*time.Millisecond)

	ca.Authenticate("user1", "pass", []string{})
	assert.Equal(t, 1, len(auther.calls))

	ca.Authenticate("user1", "pass", []string{})
	assert.Equal(t, 1, len(auther.calls)) // still 1, because it's still cached

	time.Sleep(1 * time.Second)
	ca.Authenticate("user1", "pass", []string{})
	assert.Equal(t, 2, len(auther.calls)) // and now it's 2!
}

func TestCacheFailedAuth(t *testing.T) {
	auther := testAuther(AuthFailed)
	ca := NewCachedAuthenticator(auther, 2*time.Second)

	ca.Authenticate("user1", "pass", []string{})
	ca.Authenticate("user1", "pass", []string{})
	assert.Equal(t, 2, len(auther.calls)) // failures aren't cached, so we should have 2
}
