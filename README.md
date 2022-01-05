# goauth

Designed to work with Traefik's forwardauth - <https://doc.traefik.io/traefik/v2.0/middlewares/forwardauth/>

## Application
### Configuration
Environment variables:
| Variable | Type | Default | Description |
|-|-|-|-|
| `GOAUTH_PORT` | int | `8080` | Listening port
| `GOAUTH_REALM` | string | `goauth` | HTTP basic auth realm
| `GOAUTH_LOGLEVEL` | string | `info` | Loglevel - see <https://github.com/sirupsen/logrus/blob/master/logrus.go#L25>
| `GOAUTH_LDAP_TLS` | bool | `true` | Wether to use TLS when connecting to LDAP server
| `GOAUTH_LDAP_HOSTNAME` | string | | LDAP server hostname
| `GOAUTH_LDAP_PORT` | int | `389`/`636` | LDAP server port number
| `GOAUTH_LDAP_BASEDN` | string | | User base dn (e.g `ou=users,dc=myorg,dc=com`)
| `GOAUTH_LDAP_FILTER` | string | `(uid=%s)` | User search filter
| `GOAUTH_LDAP_NAMEATTRIBUTE` | string | `uid` | Username attribute
| `GOAUTH_LDAP_GROUP_DN` | string | | Group base dn (e.g `ou=groups,dc=myorg,dc=com`)
| `GOAUTH_LDAP_GROUP_FILTER` | string | `(cn=%s)` | Group filter
| `GOAUTH_LDAP_GROUP_MEMBERATTRIBUTE` | string | `memberUid` | Membership attribute in group
| `GOAUTH_GROUPS` | []string | | Comma-separated list of required groups for all users

### Example docker-compose
```yaml
version: "3"

services:
    goauth:
        image: goauth
        container_name: goauth
        hostname: goauth
        restart: unless-stopped
        ports:
            - 8080:8080
        environment:
            - GOAUTH_REALM=Local network
            - GOAUTH_LDAP_HOSTNAME=ldap.server.com
            - GOAUTH_LDAP_BASEDN=ou=people,dc=server,dc=com
            - GOAUTH_LDAP_GROUP_DN=ou=groups,dc=server,dc=com
        volumes:
            - path/to/myCA.crt:/etc/ssl/certs/myCA.pem
```

## Usage
### Authentication request
`GET /auth?[groups=group1,group2,...]`

#### Parameters
| Type | Name | Description |
|-|-|-|
| Header | `Authorization` | Regular basic auth header
| Query | `groups` | Comma-separated list of required groups (in addition to globally configured required groups)

### Example - failure
```
GET /auth?groups=a,b HTTP/1.1
Host: localhost:8080
Authorization: Basic dXNlcjpwYXNz
Accept: */*

HTTP/1.1 401 Unauthorized
Www-Authenticate: Basic realm="realm"
Date: Fri, 25 Jun 2021 14:12:24 GMT
Content-Length: 12
Content-Type: text/plain; charset=utf-8

Unauthorized
```

### Example - success
```
GET /auth?groups=a,b HTTP/1.1
Host: localhost:8080
Authorization: Basic dXNlcjpwYXNz
Accept: */*

HTTP/1.1 200 OK
Www-Authenticate: Basic realm="realm"
Date: Fri, 25 Jun 2021 14:12:24 GMT
Content-Length: 0
```
