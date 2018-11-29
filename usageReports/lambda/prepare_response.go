//   Copyright 2018 MSolution.IO
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package lambda

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/trackit/jsonlog"
	"gopkg.in/olivere/elastic.v5"

	"github.com/trackit/trackit-server/aws/usageReports"
	"github.com/trackit/trackit-server/aws/usageReports/lambda"
	"github.com/trackit/trackit-server/errors"
)

type (

	// Structure that allow to parse ES response for costs
	ResponseCost struct {
		Accounts struct {
			Buckets []struct {
				Key       string `json:"key"`
				Instances struct {
					Buckets []struct {
						Key  string `json:"key"`
						Cost struct {
							Value float64 `json:"value"`
						} `json:"cost"`
					} `json:"buckets"`
				} `json:"instances"`
			} `json:"buckets"`
		} `json:"accounts"`
	}

	// Structure that allow to parse ES response for Lambda Monthly instances
	ResponseLambdaMonthly struct {
		Accounts struct {
			Buckets []struct {
				Instances struct {
					Hits struct {
						Hits []struct {
							Instance lambda.InstanceReport `json:"_source"`
						} `json:"hits"`
					} `json:"hits"`
				} `json:"instances"`
			} `json:"buckets"`
		} `json:"accounts"`
	}

	// Structure that allow to parse ES response for Lambda Daily instances
	ResponseLambdaDaily struct {
		Accounts struct {
			Buckets []struct {
				Dates struct {
					Buckets []struct {
						Time      string `json:"key_as_string"`
						Instances struct {
							Hits struct {
								Hits []struct {
									Instance lambda.InstanceReport `json:"_source"`
								} `json:"hits"`
							} `json:"hits"`
						} `json:"instances"`
					} `json:"buckets"`
				} `json:"dates"`
			} `json:"buckets"`
		} `json:"accounts"`
	}

	// InstanceReport has all the information of an Lambda instance report
	InstanceReport struct {
		utils.ReportBase
		Instance Instance `json:"instance"`
	}

	// Instance contains the information of an Lambda instance
	Instance struct {
		lambda.InstanceBase
		Tags  map[string]string  `json:"tags"`
		Costs map[string]float64 `json:"costs"`
	}
)

func getLambdaInstanceReportResponse(oldInstance lambda.InstanceReport) InstanceReport {
	tags := make(map[string]string, 0)
	for _, tag := range oldInstance.Instance.Tags {
		tags[tag.Key] = tag.Value
	}
	newInstance := InstanceReport{
		ReportBase: oldInstance.ReportBase,
		Instance: Instance{
			InstanceBase: oldInstance.Instance.InstanceBase,
			Tags:         tags,
			Costs:        oldInstance.Instance.Costs,
		},
	}
	return newInstance
}

// addCostToInstance adds a cost for an instance based on billing data
func addCostToInstance(instance lambda.InstanceReport, costs ResponseCost) lambda.InstanceReport {
	if instance.Instance.Costs == nil {
		instance.Instance.Costs = make(map[string]float64, 0)
	}
	for _, accounts := range costs.Accounts.Buckets {
		if accounts.Key != instance.Account {
			continue
		}
		for _, instanceCost := range accounts.Instances.Buckets {
			if strings.Contains(instanceCost.Key, instance.Instance.Name) {
				if len(instanceCost.Key) == 19 && strings.HasPrefix(instanceCost.Key, "i-") {
					instance.Instance.Costs["instance"] += instanceCost.Cost.Value
				} else {
					instance.Instance.Costs["cloudwatch"] += instanceCost.Cost.Value
				}
			}
		}
		return instance
	}
	return instance
}

// prepareResponseLambdaDaily parses the results from elasticsearch and returns an array of Lambda daily instances report
func prepareResponseLambdaDaily(ctx context.Context, resLambda *elastic.SearchResult, resCost *elastic.SearchResult) ([]InstanceReport, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	var parsedLambda ResponseLambdaDaily
	var parsedCost ResponseCost
	instances := make([]InstanceReport, 0)
	err := json.Unmarshal(*resLambda.Aggregations["accounts"], &parsedLambda.Accounts)
	if err != nil {
		logger.Error("Error while unmarshaling ES Lambda response", err)
		return nil, err
	}
	if resCost != nil {
		err = json.Unmarshal(*resCost.Aggregations["accounts"], &parsedCost.Accounts)
		if err != nil {
			logger.Error("Error while unmarshaling ES cost response", err)
		}
	}
	for _, account := range parsedLambda.Accounts.Buckets {
		var lastDate = ""
		for _, date := range account.Dates.Buckets {
			if date.Time > lastDate {
				lastDate = date.Time
			}
		}
		for _, date := range account.Dates.Buckets {
			if date.Time == lastDate {
				for _, instance := range date.Instances.Hits.Hits {
					instance.Instance = addCostToInstance(instance.Instance, parsedCost)
					instances = append(instances, getLambdaInstanceReportResponse(instance.Instance))
				}
			}
		}
	}
	return instances, nil
}

// prepareResponseLambdaMonthly parses the results from elasticsearch and returns an array of Lambda monthly instances report
func prepareResponseLambdaMonthly(ctx context.Context, resLambda *elastic.SearchResult) ([]InstanceReport, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	var response ResponseLambdaMonthly
	instances := make([]InstanceReport, 0)
	err := json.Unmarshal(*resLambda.Aggregations["accounts"], &response.Accounts)
	if err != nil {
		logger.Error("Error while unmarshaling ES Lambda response", err)
		return nil, errors.GetErrorMessage(ctx, err)
	}
	for _, account := range response.Accounts.Buckets {
		for _, instance := range account.Instances.Hits.Hits {
			instances = append(instances, getLambdaInstanceReportResponse(instance.Instance))
		}
	}
	return instances, nil
}
