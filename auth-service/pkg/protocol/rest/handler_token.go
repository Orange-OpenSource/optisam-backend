package rest

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

const (
	NORECORDS      = "no records found"
	USER_NOT_FOUND = "USER_NOT_FOUND"
	GENERICERROR   = "INTERNAL_SERVER_ERROR"
)

func (h *handler) token(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logger.Log.Info("Handler.token", zap.Any("before auth token creation", time.Now()))
	if err := h.oauth2Server.HandleTokenRequest(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	logger.Log.Info("Handler.token", zap.Any("after auth token creation", time.Now()))
}

func (h *handler) activateAccount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	h.checkandredirect(w, r, "activation")
	return
}

func (h *handler) resetPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	h.checkandredirect(w, r, "resetPassword")
	return
}

func (h *handler) setPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseForm(); err != nil {
		logger.Log.Sugar().Errorw("auth-handler - setPassword - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		sendResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req v1.ChangePasswordRequest
	err := decoder.Decode(&req)
	if err != nil {
		logger.Log.Sugar().Errorw("auth/handler - set password - error while decoding body - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		sendResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	if req.Action == "0" {
		req.TokenType = "activation"
	} else if req.Action == "1" {
		req.TokenType = "resetPassword"
	} else {
		return
	}
	err = h.service.ChangePassword(r.Context(), &req)
	if err != nil {
		logger.Log.Sugar().Errorw("auth-handler - change password - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		sendResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}
	logger.Log.Sugar().Info("auth-handler - change password - success",
		"status", codes.OK,
	)
	sendResponse(http.StatusOK, "Password have been successfully updated", w)
	return
}

func (h *handler) forgotPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseForm(); err != nil {
		logger.Log.Sugar().Errorw("auth/handler - forgot password "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		sendResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var req v1.ForgotPasswordRequest
	err := decoder.Decode(&req)
	if err != nil {
		logger.Log.Sugar().Errorw("auth/handler - forgot password - error while decoding body - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		sendResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}
	err = h.service.ForgotPassword(r.Context(), req.Username)
	if err != nil {
		logger.Log.Sugar().Errorw("auth/handler - forgot password - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		code := http.StatusInternalServerError
		returnErr := GENERICERROR
		if strings.Contains(err.Error(), NORECORDS) {
			code = http.StatusBadRequest
			returnErr = USER_NOT_FOUND
		}
		sendResponse(code, returnErr, w)
		return
	}
	logger.Log.Sugar().Info("auth/handler - forgot password - success ",
		"status", codes.OK,
		"reason", "Password ResetLink Have been sent to your email",
	)
	sendResponse(http.StatusOK, "Password ResetLink Have been sent to your email", w)
	return
}

func (h *handler) checkandredirect(w http.ResponseWriter, r *http.Request, tokenType string) {
	userId := r.URL.Query().Get("user")
	token := r.URL.Query().Get("token")
	queryParams := url.Values{}
	req := &v1.TokenRequest{
		Username:  userId,
		Token:     token,
		TokenType: tokenType,
	}
	err := h.service.TokenValidation(r.Context(), req)
	if err != nil {
		logger.Log.Sugar().Errorw("auth/handler - check and redirect - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		queryParams.Add("code", "0401")
		queryParams.Add("error", err.Error())
	} else {
		queryParams.Add("code", "0200")
		queryParams.Add("token", token)
		queryParams.Add("user", userId)
	}
	// Build the URL for the error page with the query parameter
	var redirectURL string
	if tokenType == "activation" {
		redirectURL = h.cfg.Emailtemplate.Redirectappactivation + queryParams.Encode()
	} else if tokenType == "resetPassword" {
		redirectURL = h.cfg.Emailtemplate.Redirectappurlforgotpass + queryParams.Encode()
	}
	logger.Log.Sugar().Errorw("auth/handler - setandredirect - success ",
		"status", codes.OK,
		"reason", "success",
	)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	return
}
