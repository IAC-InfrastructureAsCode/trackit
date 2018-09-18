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

package history

import (
	"time"
	"context"
	"errors"
	"net/http"
	"encoding/json"

	"github.com/trackit/jsonlog"
	"gopkg.in/olivere/elastic.v5"

	"github.com/trackit/trackit-server/aws"
	"github.com/trackit/trackit-server/es"
	"github.com/trackit/trackit-server/aws/ec2"
)

type (
	EsInstancePerProductResult struct {
		Products struct {
			Buckets []struct {
				Product   string                  `json:"key"`
				Instances EsCostPerInstanceResult `json:"instances"`
			} `json:"buckets"`
		} `json:"products"`
	}

	EsCostPerInstanceResult struct {
		Buckets []struct {
			Instance string `json:"key"`
			Cost     struct {
				Value float64 `json:"value"`
			} `json:"cost"`
		} `json:"buckets"`
	}

	CostPerInstance struct {
		Instance string
		Cost     float64
	}

	InstancePerProduct struct {
		Product   string
		Instances []CostPerInstance
	}

	Response []InstancePerProduct
)

func getHistoryDate() (time.Time, time.Time) {
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), 0, 23, 59, 59, 9999, now.Location())
	start := time.Date(now.Year(), now.Month() - 1, 0, 24, 0, 0, 0, now.Location())
	return start, end
}

// makeElasticSearchRequestForCost will make the actual request to the ElasticSearch
// It will return the data, an http status code (as int) and an error.
// Because an error can be generated, but is not critical and is not needed to be known by
// the user (e.g if the index does not exists because it was not yet indexed ) the error will
// be returned, but instead of having a 500 status code, it will return the provided status code
// with empty data
func makeElasticSearchRequestForCost(ctx context.Context, client *elastic.Client, aa aws.AwsAccount,
	startDate time.Time, endDate time.Time) (*elastic.SearchResult, int, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	index := es.IndexNameForUserId(aa.UserId, es.IndexPrefixLineItems)
	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewTermQuery("usageAccountId", es.GetAccountIdFromRoleArn(aa.RoleArn)))
	query = query.Filter(elastic.NewTermsQuery("productCode", "AmazonEC2", "AmazonRDS"))
	query = query.Filter(elastic.NewRangeQuery("usageStartDate").
		From(startDate.Format("2006-01-02T15:04:05Z")).To(endDate.Format("2006-01-02T15:04:05Z")))
	search := client.Search().Index(index).Size(0).Query(query)
	search.Aggregation("products", elastic.NewTermsAggregation().Field("productCode").
		SubAggregation("instances", elastic.NewTermsAggregation().Field("resourceId").
			SubAggregation("cost", elastic.NewSumAggregation().Field("unblendedCost"))))
	result, err := search.Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			logger.Warning("Query execution failed, ES index does not exists", map[string]interface{}{"index": index, "error": err.Error()})
			return nil, http.StatusOK, err
		}
		logger.Error("Query execution failed", err.Error())
		return nil, http.StatusInternalServerError, nil
	}
	return result, http.StatusOK, nil
}

func getEc2AndRdsCostPerInstance(ctx context.Context, aa aws.AwsAccount, startDate time.Time, endDate time.Time) (Response, error) {
	var parsedResult EsInstancePerProductResult
	var response     Response
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	result, returnCode, err := makeElasticSearchRequestForCost(ctx, es.Client, aa, startDate, endDate)
	if err != nil {
		if returnCode != http.StatusOK {
			return nil, err
		} else {
			return nil, nil
		}
	}
	err = json.Unmarshal(*result.Aggregations["products"], &parsedResult.Products)
	if err != nil {
		logger.Error("Error while unmarshaling", err)
		return nil, errors.New("Internal server error")
	}
	for _, product := range parsedResult.Products.Buckets {
		res := InstancePerProduct{product.Product, []CostPerInstance{}}
		for _, instance := range product.Instances.Buckets {
			res.Instances = append(res.Instances, CostPerInstance{instance.Instance, instance.Cost.Value})
		}
		response = append(response, res)
	}
	return response, nil
}

// checkAlreadyHistory checks if there is already an history report in ES.
// If there is already one it returns true, otherwise it returns false
func checkAlreadyHistory(ctx context.Context, date time.Time, aa aws.AwsAccount, prefix string) (bool, error) {
	return false, nil
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewTermQuery("account", es.GetAccountIdFromRoleArn(aa.RoleArn)))
	query = query.Filter(elastic.NewTermQuery("reportDate", date.Format("2006-01-02T15:04:05Z")))
	index := es.IndexNameForUserId(aa.UserId, prefix)
	result, err := es.Client.Search().Index(index).Size(0).Query(query).Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			logger.Warning("Query execution failed, ES index does not exists", map[string]interface{}{"index": index, "error": err.Error()})
			return false, nil
		}
		logger.Error("Query execution failed", err.Error())
		return false, err
	}
	if result.Hits.TotalHits == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// checkBillingDataCompleted checks if billing data in ES are complete.
// If they are complete it returns true, otherwise it returns false
func checkBillingDataCompleted(ctx context.Context, startDate time.Time, endDate time.Time, aa aws.AwsAccount) (bool, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewTermQuery("usageAccountId", es.GetAccountIdFromRoleArn(aa.RoleArn)))
	query = query.Filter(elastic.NewTermQuery("invoiceId", ""))
	query = query.Filter(elastic.NewRangeQuery("usageStartDate").
		From(startDate.Format("2006-01-02T15:04:05Z")).To(endDate.Format("2006-01-02T15:04:05Z")))
	index := es.IndexNameForUserId(aa.UserId, es.IndexPrefixLineItems)
	result, err := es.Client.Search().Index(index).Size(0).Query(query).Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			logger.Warning("Query execution failed, ES index does not exists", map[string]interface{}{"index": index, "error": err.Error()})
			return false, nil
		}
		logger.Error("Query execution failed", err.Error())
		return false, err
	}
	if result.Hits.TotalHits == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func getInstancesInfo(ctx context.Context, aa aws.AwsAccount, response Response, startDate time.Time, endDate time.Time) (error) {
	var errorString string
	for _, product := range response {
		switch product.Product {
		case "AmazonEC2":
			if already, err := checkAlreadyHistory(ctx, startDate, aa, ec2.IndexPrefixEC2Report); already || err != nil {
			} else if report, err := getEc2InstancesInfos(ctx, product.Instances, aa); err != nil {
				errorString += "Error while fetching EC2 infos. "
			} else if err = putEc2ReportInES(ctx, report, aa); err != nil {
				errorString += "Error while saving EC2 report in ES. "
			}
		}
	}
	if errorString != "" {
		return errors.New(errorString)
	} else {
		return nil
	}
}

func FetchHistoryInfos(ctx context.Context, aa aws.AwsAccount) (error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	startDate, endDate := getHistoryDate()
	logger.Info("Fetching EC2 and RDS history for " + string(aa.Id) + " (" + aa.Pretty + ")", map[string]interface{}{
		"awsAccountId": aa.Id,
		"startDate":    startDate.Format("2006-01-02T15:04:05Z"),
		"endDate":      endDate.Format("2006-01-02T15:04:05Z"),
	})
	if complete, err := checkBillingDataCompleted(ctx, startDate, endDate, aa); !complete || err != nil {
		return err
	}
	response, err := getEc2AndRdsCostPerInstance(ctx, aa, startDate, endDate)
	if err != nil {
		logger.Error("Error while getting EC2 and RDS cost per instance on ES.", err.Error())
		return err
	}
	err = getInstancesInfo(ctx, aa, response, startDate, endDate)
	return err
}