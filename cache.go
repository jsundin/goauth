package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type CachedAuthenticator struct {
	delegate Authenticator
	ttl      time.Duration
	cache    map[string]time.Time
	mutex    *sync.Mutex
}

func NewCachedAuthenticator(delegate Authenticator, ttl time.Duration) *CachedAuthenticator {
	ca := &CachedAuthenticator{
		delegate: delegate,
		ttl:      ttl,
		cache:    make(map[string]time.Time),
		mutex:    &sync.Mutex{},
	}
	return ca
}

func (ca *CachedAuthenticator) Authenticate(user, pass string, extraGroups []string) int {
	cacheKey := ca.Key(user, pass, extraGroups)

	ca.mutex.Lock()
	defer ca.mutex.Unlock()

	log.Debugf("cache check: key=%v, cache=%v", cacheKey, ca.cache)
	if expiration, foundInCache := ca.cache[cacheKey]; foundInCache {
		if expiration.After(time.Now()) {
			log.Debugf("cache hit")
			return AuthSuccess
		}
	}

	result := ca.delegate.Authenticate(user, pass, extraGroups)

	if result == AuthSuccess {
		expiration := time.Now().Add(ca.ttl)
		ca.cache[cacheKey] = expiration
		log.Debugf("added key=%v to cache with expiration=%v", cacheKey, expiration)
	} else {
		delete(ca.cache, cacheKey)
	}

	return result
}

func (ca *CachedAuthenticator) Key(user, pass string, extraGroups []string) string {
	keyData := struct {
		U string
		P string
		G []string
	}{
		U: user,
		P: pass,
		G: extraGroups,
	}
	buffer := bytes.NewBuffer(nil)
	json.NewEncoder(buffer).Encode(keyData)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}
