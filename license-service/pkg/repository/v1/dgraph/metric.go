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

// ListMetrices implements Licence ListMetrices function
func (l *LicenseRepository) ListMetrices(ctx context.Context, scopes ...string) ([]*v1.Metric, error) {

	q := `   {
             Metrics(func:eq(type_name,"metric")) ` + agregateFilters(scopeFilters(scopes)) + `{
			   ID  : uid
			   Name: metric.name
			   Type: metric.type
		   }
		}
		  `

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("listMetrices - cannot complete query transaction")
	}

	type Data struct {
		Metrics []*v1.Metric
	}
	var metricList Data
	if err := json.Unmarshal(resp.GetJson(), &metricList); err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("listMetrices - cannot unmarshal Json object")
	}

	return metricList.Metrics, nil
}

func (l *LicenseRepository) listMetricWithMetricType(ctx context.Context, metType v1.MetricType, scopes ...string) (json.RawMessage, error) {
	q := `{
		Data(func: eq(metric.type,` + metType.String() + `)) @filter(eq(scopes,[` + strings.Join(scopes, ",") + `])) {
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/listMetricWithMetricType - query failed", zap.Error(err), zap.String("query", q))
		return nil, fmt.Errorf("cannot get metrics of %s", metType.String())
	}
	return resp.Json, nil
}

func filterMetricEquipments(prodMetricEquip []*v1.ProductAllocationEquipmentMetrics, metricName ...string) map[string]interface{} {
	resp := make(map[string]interface{})

	// Filter out allocated equipment to metric and not allocated to this metric
	// Find out the UIDs of those equipments for forcesing to calculate complaince on that equipments only
	// Find out the UIDs of those equipment which are allocated but not to this metric. These equipment will exclude from calculation
	//logger.Log.Sugar().Infow("Data for metric alloted", "metricName", metricName, "ProductEquipment", prodMetricEquip[0].ProductEquipment, "AllocatedEquipMetric", prodMetricEquip[0].MetricAllocation)
	allotedUIDs := ""
	notAllotedUIDs := ""
	notAllotedVirtualMachineUIDs := ""
	notAllocatedUserID := ""
	if len(prodMetricEquip) > 0 {
		for index := range prodMetricEquip {
			for _, equip := range prodMetricEquip[index].ProductEquipment {
				if prodMetricEquip[index].MetricAllocation != nil {
					for _, allotedEquip := range prodMetricEquip[index].MetricAllocation {
						if allotedEquip.EquipmentId == equip.EquipmentId {
							if allotedEquip.MetricAllocated == metricName[0] || (len(metricName[1]) > 0 && allotedEquip.MetricAllocated == metricName[1]) {
								if allotedUIDs != "" {
									allotedUIDs += " OR "
								}
								allotedUIDs += " uid(" + equip.EUID + ") "
							} else {
								if equip.EquipmentType == "server" {
									if notAllotedUIDs != "" {
										notAllotedUIDs += " , "
									}
									notAllotedUIDs += " " + equip.EUID + " "
								}
								if equip.EquipmentType == "virtualMachine" {
									if notAllotedVirtualMachineUIDs != "" {
										notAllotedVirtualMachineUIDs += " , "
									}
									notAllotedVirtualMachineUIDs += " " + equip.EUID + " "
								}
								if notAllocatedUserID != "" {
									notAllocatedUserID += " , "
								}
								notAllocatedUserID += "\"user_" + prodMetricEquip[index].SwidTag + "" + equip.EquipmentId + "\""

							}
						}
					}
				}
			}
		}
	}
	resp["notAllotedVirtualMachine"] = notAllotedVirtualMachineUIDs
	resp["alloted"] = allotedUIDs
	resp["notAlloted"] = notAllotedUIDs
	resp["notAllocatedUserID"] = notAllocatedUserID
	return resp
}
