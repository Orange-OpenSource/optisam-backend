package rest

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"optisam-backend/common/optisam/logger"
	"os"

	"net/http"
	"net/http/httptest"
	"net/url"
	"optisam-backend/auth-service/pkg/api/v1"
	mock_authService "optisam-backend/auth-service/pkg/api/v1/mock"
	mock_acctok "optisam-backend/auth-service/pkg/oauth2/generators/access/mock"
	optisam_oauth2Server "optisam-backend/auth-service/pkg/oauth2/server"
	mock_clientstore "optisam-backend/auth-service/pkg/oauth2/stores/client/mock"
	mock_tokenstore "optisam-backend/auth-service/pkg/oauth2/stores/token/mock"
	"strings"
	"testing"

	"gopkg.in/oauth2.v3/models"

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"

	"gopkg.in/oauth2.v3/server"
)

func TestMain(m *testing.M) {
	logger.Init(1, "")
	os.Exit(m.Run())
}

func tokenRequest(reqURL, grantType, username, password string) (*http.Request, error) {
	data := url.Values{}
	data.Set("grant_type", grantType)
	data.Set("username", username)
	data.Set("password", password)
	payload := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", reqURL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func Test_handler_token(t *testing.T) {
	var service v1.AuthService
	var srv *server.Server
	var mockCtrl *gomock.Controller
	type args struct {
		grantType string
		username  string
		password  string
	}
	tests := []struct {
		name   string
		args   args
		setup  func()
		assert func(resp *http.Response) error
	}{
		{name: "server",
			args: args{
				grantType: "password",
				username:  "user",
				password:  "secret",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockService := mock_authService.NewMockAuthService(mockCtrl)
				mockService.EXPECT().Login(gomock.Any(), &v1.LoginRequest{
					Username: "user",
					Password: "secret",
				}).Return(&v1.LoginResponse{
					UserID: "user",
				}, nil).Times(1)

				service = mockService
				mockClientStore := mock_clientstore.NewMockClientStore(mockCtrl)

				mockClientStore.EXPECT().GetByID("").Return(&models.Client{}, nil).Times(1)

				mockAccTokGen := mock_acctok.NewMockAccessGenerate(mockCtrl)
				mockAccTokGen.EXPECT().Token(gomock.Any(), true).
					Return("access", "refresh", nil).Times(1)

				mockTokenStore := mock_tokenstore.NewMockTokenStore(mockCtrl)
				mockTokenStore.EXPECT().Create(gomock.Any()).Return(nil).Times(1)

				srv = optisam_oauth2Server.NewServer(mockTokenStore, mockClientStore, mockAccTokGen)
			},
			assert: func(resp *http.Response) error {
				if http.StatusOK != resp.StatusCode {
					return fmt.Errorf("expected status code: %v OK, got: %v %v", http.StatusOK, resp.Status, resp.StatusCode)
				}
				defer resp.Body.Close()

				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				expData := `{"access_token":"access","expires_in":7200,"refresh_token":"refresh","token_type":"Bearer"}`

				if string(bytes.TrimSpace(data)) != expData {
					return fmt.Errorf("expected: %s, got: %s", expData, data)
				}

				return nil
			},
		},
		{name: "failure",
			args: args{
				grantType: "password",
				username:  "user",
				password:  "secret",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockService := mock_authService.NewMockAuthService(mockCtrl)
				mockService.EXPECT().Login(gomock.Any(), &v1.LoginRequest{
					Username: "user",
					Password: "secret",
				}).Return(nil, errors.New("test error")).Times(1)

				service = mockService
				srv = optisam_oauth2Server.NewServer(nil, nil, nil)
			},
			assert: func(resp *http.Response) error {
				if http.StatusInternalServerError != resp.StatusCode {
					return fmt.Errorf("expected status code: %v, got: %v %v", http.StatusOK, resp.Status, resp.StatusCode)
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			handler := newHandler(service, srv)
			router := httprouter.New()
			router.POST("/api/v1/token", handler.token)
			tServer := httptest.NewServer(router)
			defer tServer.Close()
			req, err := tokenRequest(tServer.URL+"/api/v1/token", tt.args.grantType, tt.args.username, tt.args.password)
			if !assert.Empty(t, err) {
				return
			}
			resp, err := tServer.Client().Do(req)
			if !assert.Empty(t, err) {
				return
			}
			if !assert.Empty(t, tt.assert(resp)) {
				return
			}
			mockCtrl.Finish()
		})
	}
}
