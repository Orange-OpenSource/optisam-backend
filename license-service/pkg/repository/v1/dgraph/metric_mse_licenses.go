package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

// MetricMSEComputedLicenses implements Licence MetricMSEComputedLicenses function
func (l *LicenseRepository) MetricMSEComputedLicenses(ctx context.Context, sa bool, id []string, mat *v1.MetricMSEComputed, scopes ...string) (uint64, error) {
	totalVMCount, totalServerCount := l.productEquipmentCount(ctx, id...)
	if totalVMCount == 0 && totalServerCount == 0 {
		return 0, nil
	}
	q := queryBuilderMSE(mat, scopes, totalVMCount, totalServerCount, sa, id...)
	prod, err := l.licensesForMSE(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorf("dgraph/MetricMSEComputedLicenses - licensesForMSE", zap.Error(err))
		return 0, errors.New("dgraph/MetricMSEComputedLicenses - query failed")
	}

	return prod, nil
}

// MetricMSEComputedLicensesAgg implements Licence MetricMSEComputedLicensesAgg function
func (l *LicenseRepository) MetricMSEComputedLicensesAgg(ctx context.Context, sa bool, name, metric string, mat *v1.MetricMSEComputed, scopes ...string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Sugar().Errorf("dgraph/MetricMSEComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricMSEComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}
	totalVMCount, totalServerCount := l.productEquipmentCount(ctx, ids...)
	if totalVMCount == 0 && totalServerCount == 0 {
		return 0, nil
	}
	q := queryBuilderMSE(mat, scopes, totalVMCount, totalServerCount, sa, ids...)
	fmt.Println(q)
	prod, err := l.licensesForMSE(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorf("dgraph/MetricMSEComputedLicensesAgg - licensesForMSE", zap.Error(err))
		return 0, errors.New("dgraph/MetricMSEComputedLicensesAgg - query failed")
	}

	return prod, nil
}

func (l *LicenseRepository) productEquipmentCount(ctx context.Context, id ...string) (int32, int32) {
	query := `{
		var(func: uid(` + strings.Join(id, ",") + `)){
			product.equipment @filter(eq(equipment.type,virtualmachine)){
				softpartitionIDs as count(equipment.id)
			}
		}
		var(func: uid(` + strings.Join(id, ",") + `)){
			product.equipment @filter(eq(equipment.type,server)){
				serverIDs as count(equipment.id)
			}
		  
		}
		ProductEquipment(){
			TotalVMCount : sum(val(softpartitionIDs))
			TotalServerCount :sum(val(serverIDs))
		}	
	}`
	resp, err := l.dg.NewTxn().Query(ctx, query)
	if err != nil {
		logger.Log.Sugar().Errorf("dgraph/productEquipmentCount - query failed", "error", err.Error(), "query", query)
		return 0, 0
	}
	type productEquipment struct {
		TotalVMCount     int32 `json:"TotalVMCount"`
		TotalServerCount int32 `json:"TotalServerCount"`
	}
	type TotalProductEquipment struct {
		ProductEquipment []*productEquipment `json:"ProductEquipment"`
	}

	data := &TotalProductEquipment{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		return 0, 0
	}
	if len(data.ProductEquipment) == 0 {
		return 0, 0
	}
	return data.ProductEquipment[0].TotalVMCount, data.ProductEquipment[1].TotalServerCount
}
func (l *LicenseRepository) licensesForMSE(ctx context.Context, q string) (uint64, error) {
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorf("dgraph/MetricMSEComputedLicenses - query failed", "error", err.Error(), "query", q)
		return 0, fmt.Errorf("query failed, err: %v", err)
	}

	type licenses struct {
		Licenses float64 `json:"Licenses"`
	}

	// type vmcomp struct {
	// 	p_cores int32   `json:"p_cores"`
	// 	p_comp  float64 `json:"p_comp"`
	// 	sumVcpu int32   `json:"sumVcpu"`
	// 	vm_comp float64 `json:"vm_comp"`
	// }

	type totalLicenses struct {
		//VmComp   []*vmcomp   `json:"VmComp"`
		Licenses []*licenses `json:"Licenses"`
	}

	data := &totalLicenses{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		return 0, fmt.Errorf("unmarshal failed, err: %v", err)
	}
	if len(data.Licenses) == 0 {
		return 0, v1.ErrNoData
	}

	// for i := range data.VmComp {
	// 	if sa && data.Licenses[0].Licenses == 0 {
	// 		data.Licenses[0].Licenses = math.Min(data.VmComp[i].p_comp, data.VmComp[i].vm_comp)
	// 	} else {
	// 		if data.VmComp[i].sumVcpu <= data.VmComp[i].p_cores {
	// 			data.Licenses[0].Licenses += math.Min(data.VmComp[i].p_comp, data.VmComp[i].vm_comp)
	// 		} else {
	// 			data.Licenses[0].Licenses += data.VmComp[i].p_comp + data.VmComp[i].vm_comp
	// 		}
	// 	}
	// }

	return uint64(data.Licenses[0].Licenses), nil
}
