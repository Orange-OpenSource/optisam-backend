package postgres

import (
	"context"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	"strings"

	"go.uber.org/zap"
)

// UpsertProductTx upserts products/ linking data

// nolint: gosec
func (p *DpsRepository) StoreCoreFactorReferences(ctx context.Context, data map[string]map[string]string) error {
	if len(data) == 0 {
		return errors.New("emptyReference")
	}

	rows := ""
	index := 1
	for k, v := range data {
		for x, y := range v {
			rows += fmt.Sprintf("(%d,'%s','%s','%s'),", index, k, x, y)
			index++
		}
	}
	rows = strings.TrimRight(rows, ",")
	query := fmt.Sprintf("insert into core_factor_references values %s on conflict (id) do update set manufacturer=EXCLUDED.manufacturer , model=EXCLUDED.model , core_factor= EXCLUDED.core_factor;", rows)
	logger.Log.Debug("Batch Query", zap.String("query", query))

	if _, err := p.db.Exec(query); err != nil {
		logger.Log.Error("Batch insertion failure for core factor reference", zap.Error(err))
		return err
	}
	logger.Log.Debug("Batch inserted ", zap.Any("Batch", query))
	return nil
}
