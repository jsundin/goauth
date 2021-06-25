package main

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap"
	log "github.com/sirupsen/logrus"
)

type LdapAuthenticatorConfig struct {
	TLS      bool   `default:"true"`
	Hostname string `required:"true"`
	Port     *int
	Insecure bool `default:"false"`

	BaseDN        string
	Filter        string `default:"(uid=%s)"`
	NameAttribute string `default:"uid"`

	Group struct {
		DN              string
		Filter          string `default:"(cn=%s)"`
		MemberAttribute string `default:"memberUid"`
	}

	Groups []string
}

func (conf *LdapAuthenticatorConfig) Authenticate(user, pass string, extraGroups []string) bool {
	conn, err := conf.getConnection()
	if err != nil {
		log.Errorf("could not connect to ldap server: %v", err)
		return false
	}
	defer conn.Close()

	ldapUser, err := conf.findOneItem(conn, conf.BaseDN, strings.ReplaceAll(conf.Filter, "%s", user), ldap.ScopeWholeSubtree)
	if err != nil {
		log.Debugf("could not find user '%v': %v", user, err)
		return false
	}
	uid := ldapUser.GetAttributeValue(conf.NameAttribute)
	if uid != user {
		log.Errorf("user '%s' does not have required attribute '%v' with the username (actual value: '%v')", user, conf.NameAttribute, uid)
	}

	err = conf.checkRequiredGroups(conn, user, conf.Groups)
	if err != nil {
		log.Debugf("user '%s' does not have all required global groups: %v", user, err)
		return false
	}

	err = conf.checkRequiredGroups(conn, user, extraGroups)
	if err != nil {
		log.Debugf("user '%s' does not have all required local groups: %v", user, err)
		return false
	}

	err = conn.Bind(ldapUser.DN, pass)
	if err != nil {
		log.Debugf("bind failed for user '%v': %v", ldapUser.DN, err)
		return false
	}

	return true
}

func (conf *LdapAuthenticatorConfig) getConnection() (*ldap.Conn, error) {
	var conn *ldap.Conn
	var err error

	if conf.TLS {
		port := 636
		if conf.Port != nil {
			port = *conf.Port
		}

		insecure := false
		if conf.Insecure {
			insecure = true
		}

		conn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", conf.Hostname, port), &tls.Config{InsecureSkipVerify: insecure})
	} else {
		port := 389
		if conf.Port != nil {
			port = *conf.Port
		}

		conn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", conf.Hostname, port))
	}

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (conf *LdapAuthenticatorConfig) findOneItem(conn *ldap.Conn, baseDN, filter string, scope int) (*ldap.Entry, error) {
	request := &ldap.SearchRequest{
		BaseDN: baseDN,
		Filter: filter,
		Scope:  scope,
	}
	log.Debugf("ldap search: %v", request)
	result, err := conn.Search(request)
	if err != nil {
		log.Errorf("ldap search failed: %v", err)
		return nil, err
	}
	if len(result.Entries) != 1 {
		if len(result.Entries) > 1 {
			return nil, fmt.Errorf("found %d values, expected only one", len(result.Entries))
		}
		return nil, fmt.Errorf("not found")
	}

	return result.Entries[0], nil
}

func (conf *LdapAuthenticatorConfig) checkRequiredGroups(conn *ldap.Conn, user string, groups []string) error {
	for _, group := range groups {
		ldapGroup, err := conf.findOneItem(conn, conf.Group.DN, strings.ReplaceAll(conf.Group.Filter, "%s", group), ldap.ScopeSingleLevel)
		if err != nil {
			return fmt.Errorf("could not find group '%v': %v", group, err)
		}

		members := ldapGroup.GetAttributeValues(conf.Group.MemberAttribute)
		found := false
		for _, member := range members {
			if member == user {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("user '%v' is not a member of '%v'", user, group)
		}
	}
	return nil
}
