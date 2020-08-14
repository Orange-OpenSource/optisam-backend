// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *handler) token(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := h.oauth2Server.HandleTokenRequest(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
