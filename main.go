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

type Authenticator interface {
	Authenticate(user, pass string, extraGroups []string) bool
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
		log.Debugf("request", r)

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
		if !hasAuth || !auther.Authenticate(user, pass, extraGroups) {
			rw.Header().Set("WWW-Authenticate", "Basic realm=\""+conf.Realm+"\"")
			rw.WriteHeader(401)
			rw.Write([]byte("Unauthorized"))
			return
		}
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil); err != nil {
		log.Panic(err)
	}
}
