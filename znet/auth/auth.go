package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

const UserKey = "auth_user"

type (
	Accounts  map[string]string
	authPairs []authPair
	authPair  struct {
		value string
		user  string
	}
)

func (a authPairs) searchCredential(authValue string) (string, bool) {
	if authValue == "" {
		return "", false
	}

	for _, pair := range a {
		if subtle.ConstantTimeCompare(zstring.String2Bytes(pair.value), zstring.String2Bytes(authValue)) == 1 {
			return pair.user, true
		}
	}
	return "", false
}

func New(accounts Accounts) znet.Handler {
	return BasicRealm(accounts, "")
}

func BasicRealm(accounts Accounts, realm string) znet.Handler {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)
	pairs, err := processAccounts(accounts)
	zerror.Panic(err)

	return func(c *znet.Context) {
		user, found := pairs.searchCredential(c.GetHeader("Authorization"))
		if !found {
			c.SetHeader("WWW-Authenticate", realm)
			c.Abort(http.StatusUnauthorized)
			return
		}

		c.WithValue(UserKey, user)
		c.Next()
	}
}

func processAccounts(accounts Accounts) (authPairs, error) {
	length := len(accounts)
	if length == 0 {
		return nil, errors.New("empty list of authorized credentials")
	}
	pairs := make(authPairs, 0, length)
	for user, password := range accounts {
		if user == "" {
			return nil, errors.New("user can not be empty")
		}
		value := authorizationHeader(user, password)
		pairs = append(pairs, authPair{
			value: value,
			user:  user,
		})
	}
	return pairs, nil
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString(zstring.String2Bytes(base))
}
