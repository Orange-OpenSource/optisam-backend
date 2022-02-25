package handler

import "net/http"

// ClientInfoHandler returns client info like id and secret.
func ClientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
	return "", "", nil
}
