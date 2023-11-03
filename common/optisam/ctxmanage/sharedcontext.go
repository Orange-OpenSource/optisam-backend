package ctxmanage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"google.golang.org/grpc/metadata"
)

const (
	failresmap = "failed to get respmap data"
	PASSWORD   = "password"
	BASEURLSTR = "http://%s/api/v1/token"
)

func CreateSharedContext(uname, pass, add string) (*context.Context, error) {
	logger.Log.Sugar().Infow("CreateSharedContext", "CreateSharedContext called", time.Now())
	ctx := context.Background()
	respMap := make(map[string]interface{})
	data := url.Values{
		"username":   {uname},
		"password":   {pass},
		"grant_type": {PASSWORD},
	}
	logger.Log.Sugar().Infow("CreateSharedContext", "before auth api", time.Now())
	api := fmt.Sprintf(BASEURLSTR, add)

	resp, err := http.PostForm(api, data) // nolint: gosec
	logger.Log.Sugar().Infow("CreateSharedContext", "after auth api", time.Now())
	if err != nil {
		logger.Log.Sugar().Errorw("Failed to get user claims", "err", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(bodyBytes, &respMap)
	if err != nil {
		logger.Log.Sugar().Errorw("failed to unmarshal byte data", "err", err.Error())
		return nil, err
	}
	if respMap["access_token"] == nil {
		logger.Log.Sugar().Errorw(failresmap, "error", failresmap)
		return nil, fmt.Errorf(failresmap)
	}
	authStr := fmt.Sprintf("Bearer %s", respMap["access_token"].(string))
	md := metadata.Pairs("Authorization", authStr)

	ctx = metadata.NewIncomingContext(ctx, md)

	logger.Log.Sugar().Infow("CreateSharedContext", "CreateSharedContext executed", time.Now())
	return &ctx, nil
}
