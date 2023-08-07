package main

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
)

type ctxKey struct{}

func authorizationBearerToken(ctx context.Context, token string) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, token), nil
}

func userLoggedInWithValidCredentials(ctx context.Context, credential string) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, credential), nil
}

func page_numIs(ctx context.Context, page_num int) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, page_num), nil
}

func page_sizeIs(ctx context.Context, page_size int) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, page_size), nil
}

func requestIsReports(ctx context.Context, reports string) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, reports), nil
}

func scopeIsMON(ctx context.Context, scope string) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, scope), nil
}

func sort_byIsCreated_on(ctx context.Context, sort_by string) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, sort_by), nil
}

func sort_orderIsDesc(ctx context.Context, sort_order string) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, sort_order), nil
}

func statusCodeOK(ctx context.Context, status string) error {
	return nil
}

func forEachRecordEditorIsNotNULL(ctx context.Context) error {
	return nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	reqParam := map
	context.WithValue(ctx, ctxKey{},)
	ctx.Step(`^authorization bearer token$`, authorizationBearerToken)
	ctx.Step(`^for each record Editor is not NULL$`, forEachRecordEditorIsNotNULL)
	ctx.Step(`^page_num is (\d+)$`, page_numIs)
	ctx.Step(`^page_size is (\d+)$`, page_sizeIs)
	ctx.Step(`^request is reports$`, requestIsReports)
	ctx.Step(`^scope is MON$`, scopeIsMON)
	ctx.Step(`^sort_by is created_on$`, sort_byIsCreated_on)
	ctx.Step(`^sort_order is desc$`, sort_orderIsDesc)
	ctx.Step(`^Status Code: (\d+) OK$`, statusCodeOK)
	ctx.Step(`^User logged in with valid credentials$`, userLoggedInWithValidCredentials)
}
