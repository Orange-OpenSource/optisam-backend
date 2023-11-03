package redis

import (
	"context"
	"errors"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
)

const (
	AccountCreation_ = "AccountCreation_"
	ResetPassword_   = "ResetPassword_"
	activation       = "activation"
	resetPassword    = "resetPassword"
)

func SetToken(ep helper.EmailParams, ctx context.Context, redis *redis.Client, ttl int) error {
	// set key-value pair with TTL
	var key string
	if ep.TokenType == activation {
		key = AccountCreation_ + ep.Email
	} else if ep.TokenType == resetPassword {
		key = ResetPassword_ + ep.Email
	} else {
		return errors.New("invalid tokenType")
	}

	err := redis.Set(ctx, key, ep.Token, time.Duration(ttl)*time.Minute).Err()
	if err != nil {
		logger.Log.Sugar().Errorw("common-mailer - SetToken - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return err
	}
	return nil
}

func GetToken(emaiReq helper.EmailParams, ctx context.Context, redis *redis.Client) error {
	var key string
	if emaiReq.TokenType == activation {
		key = AccountCreation_ + emaiReq.Email
	} else if emaiReq.TokenType == resetPassword {
		key = ResetPassword_ + emaiReq.Email
	} else {
		logger.Log.Sugar().Errorw("common-redis - GetToken - invalid tokenType",
			"status", codes.Internal,
			"reason", "invalid tokenType",
		)
		return errors.New("invalid tokenType")
	}
	result, err := redis.Get(ctx, key).Result()
	if err != nil {
		logger.Log.Sugar().Errorw("common-redis - GetToken - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return errors.New("token not found")
	}
	if result != emaiReq.Token {
		logger.Log.Sugar().Errorw("common-redis - GetToken - wrong token value",
			"status", codes.Internal,
			"reason", "wrong token value",
		)
		return errors.New("wrong token value")
	}
	return nil
}

func DelToken(emaiReq helper.EmailParams, ctx context.Context, redis *redis.Client) error {
	var key string
	if emaiReq.TokenType == activation {
		key = AccountCreation_ + emaiReq.Email
	} else if emaiReq.TokenType == resetPassword {
		key = ResetPassword_ + emaiReq.Email
	} else {
		return errors.New("invalid tokenType")
	}
	err := redis.Del(ctx, key).Err()
	if err != nil {
		logger.Log.Sugar().Errorw("common-redis - GetToken - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return err
	}
	return nil
}
