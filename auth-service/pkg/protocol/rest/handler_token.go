package rest

import (
	"net/http"
	"optisam-backend/common/optisam/logger"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

func (h *handler) token(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logger.Log.Info("Handler.token", zap.Any("before auth token creation", time.Now()))
	if err := h.oauth2Server.HandleTokenRequest(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	logger.Log.Info("Handler.token", zap.Any("after auth token creation", time.Now()))
}
