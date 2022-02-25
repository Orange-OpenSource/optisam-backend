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
