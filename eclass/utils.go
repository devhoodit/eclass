package eclass

import (
	"errors"
	"net/http"
)

func parseCookie(name string, cs []*http.Cookie) (string, error) {
	for _, cookie := range cs {
		if cookie.Name == name {
			return cookie.Value, nil
		}
	}
	return "", errors.New("can't find matched cookie")
}
