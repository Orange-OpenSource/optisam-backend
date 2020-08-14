// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package opa

import (
	"context"
	"encoding/json"
	"fmt"
	"optisam-backend/common/optisam/logger"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/util"
	"go.uber.org/zap"
)

type AuthzInput struct {
	MethodFullName string `json:"api"`
	Role           string `json:"role"`
}

func NewOPA(ctx context.Context, regoFile string) (*rego.PreparedEvalQuery, error) {
	regoPaths := []string{regoFile}
	logger.Log.Info("regpopath:", zap.String("regopath", regoPaths[0]))
	r, err := rego.New(rego.Query("data.rbac.allow"), rego.Load(regoPaths, nil)).PrepareForEval(ctx)
	if err != nil {
		logger.Log.Error("Failed to Load OPA Policies", zap.Error(err))
		return nil, err
	}
	return &r, nil
}

func EvalAuthZ(ctx context.Context, p *rego.PreparedEvalQuery, authzInput AuthzInput) (bool, error) {
	var input map[string]interface{}
	bs, err := json.Marshal(authzInput)
	if err != nil {
		return false, err
	}
	err = util.UnmarshalJSON(bs, &input)
	if err != nil {
		return false, err
	}

	inputValue, err := ast.InterfaceToValue(input)
	if err != nil {
		return false, err
	}

	rs, err := p.Eval(ctx, rego.EvalParsedInput(inputValue))
	if err != nil {
		return false, err
	}
	authorized := false
	switch decision := rs[0].Expressions[0].Value.(type) {
	case bool:
		if decision {
			authorized = true
		}

	default:
		err = fmt.Errorf("illegal value for policy evaluation result: %T", decision)
		return false, err
	}
	return authorized, nil
}
