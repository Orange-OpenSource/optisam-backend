// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package handler

import "net/http"

// ClientInfoHandler returns client info like id and secret.
func ClientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
	return "", "", nil
}
