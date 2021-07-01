package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	Port     int    `default:"8080"`
	Realm    string `default:"goauth"`
	LogLevel string `default:"info"`
}

const (
	AuthSuccess = iota
	AuthFailed
	AuthError
)

type Authenticator interface {
	Authenticate(user, pass string, extraGroups []string) int
}

func authenticationHandler(auther Authenticator, realm string, rw http.ResponseWriter, r *http.Request) {
	log.Debug("request", r)

	var extraGroups []string
	if requestedGroups, found := r.URL.Query()["groups"]; found {
		if len(requestedGroups) != 1 {
			rw.WriteHeader(400)
			rw.Write([]byte("Missing 'groups' parameter"))
			return
		}
		extraGroups = strings.Split(requestedGroups[0], ",")
	}

	user, pass, hasAuth := r.BasicAuth()
	var authResult int

	if hasAuth {
		authResult = auther.Authenticate(user, pass, extraGroups)
		if authResult == AuthError {
			rw.WriteHeader(500)
			rw.Write([]byte("Internal server error"))
			return
		}
	}

	if !hasAuth || authResult != AuthSuccess {
		rw.Header().Set("WWW-Authenticate", "Basic realm=\""+realm+"\"")
		rw.WriteHeader(401)
		rw.Write([]byte("Unauthorized"))
		return
	}
}

func main() {
	conf := &AppConfig{}
	if err := envconfig.Process("goauth", conf); err != nil {
		log.Panic(err)
	}
	if logLevel, err := log.ParseLevel(conf.LogLevel); err != nil {
		log.Panic(err)
	} else {
		log.SetLevel(logLevel)
	}

	ldapConf := &LdapAuthenticatorConfig{}
	if err := envconfig.Process("goauth_ldap", ldapConf); err != nil {
		log.Panic(err)
	}

	var auther Authenticator = ldapConf
	http.HandleFunc("/auth", func(rw http.ResponseWriter, r *http.Request) {
		authenticationHandler(auther, conf.Realm, rw, r)
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil); err != nil {
		log.Panic(err)
	}
}
