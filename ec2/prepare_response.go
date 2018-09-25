//   Copyright 2017 MSolution.IO
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

package ec2

import (
	"fmt"
	"strings"
	"context"
	"encoding/json"
	"database/sql"

	"gopkg.in/olivere/elastic.v5"
	"github.com/trackit/jsonlog"

	"github.com/trackit/trackit-server/users"
	"github.com/trackit/trackit-server/es"
)

type (

	// Structure that allow to parse ES response for costs
	ResponseCost struct {
		Instances struct {
			Buckets []struct {
				Key  string `json:"key"`
				Cost struct {
					Value float64 `json:"value"`
				} `json:"cost"`
			} `json:"buckets"`
		} `json:"instances"`
	}

	// Structure that allow to parse ES response for EC2 usage report
	ResponseEc2 struct {
		TopReports struct {
			Buckets []struct {
				TopReportsHits struct {
					Hits struct {
						Hits []struct {
							Source Report `json:"_source"`
						} `json:"hits"`
					} `json:"hits"`
				} `json:"top_reports_hits"`
			} `json:"buckets"`
		} `json:"top_reports"`
	}

	// Report format for EC2 usage
	Report struct {
		Account    string `json:"account"`
		ReportDate string `json:"reportDate"`
		Instances  []struct {
			Id         string             `json:"id"`
			Region     string             `json:"region"`
			State      string             `json:"state"`
			Purchasing string             `json:"purchasing"`
			Cost       float64            `json:"cost"`
			CpuAverage float64            `json:"cpuAverage"`
			CpuPeak    float64            `json:"cpuPeak"`
			NetworkIn  int64              `json:"networkIn"`
			NetworkOut int64              `json:"networkOut"`
			IORead     map[string]float64 `json:"ioRead"`
			IOWrite    map[string]float64 `json:"ioWrite"`
			KeyPair    string             `json:"keyPair"`
			Type       string             `json:"type"`
			Tags       interface{}        `json:"tags"`
		} `json:"instances"`
	}
)

// makeElasticSearchCostRequests prepares and run the request to retrieve the cost per instance
// It will return the data, an http status code (as int) and an error.
// Because an error can be generated, but is not critical and is not needed to be known by
// the user (e.g if the index does not exists because it was not yet indexed ) the error will
// be returned, but instead of having a 500 status code, it will return the provided status code
// with empty data
func makeElasticSearchCostRequest(ctx context.Context, user users.User, tx *sql.Tx, account string , date string) (ResponseCost, error) {
	l := jsonlog.LoggerFromContextOrDefault(ctx)
	accountsAndIndexes, _, err := es.GetAccountsAndIndexes([]string{account}, user, tx, es.IndexPrefixLineItems)
	if err != nil {
		return ResponseCost{}, err
	}
	index := strings.Join(accountsAndIndexes.Indexes, ",")
	searchService := GetElasticSearchCostParams(
		account,
		date,
		es.Client,
		index,
	)
	res, err := searchService.Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			l.Warning("Query execution failed, ES index does not exists : "+index, err)
			return ResponseCost{}, err
		}
		l.Error("Query execution failed : "+err.Error(), nil)
		return ResponseCost{}, fmt.Errorf("could not execute the ElasticSearch query")
	}
	var resCost ResponseCost
	err = json.Unmarshal(*res.Aggregations["instances"], &resCost.Instances)
	if err != nil {
		return ResponseCost{}, err
	}
	return resCost,  nil
}

// addCostToReport adds cost for each instance based on billing data
func addCostToReport(report Report, costs ResponseCost) (Report) {
	for _, instance := range costs.Instances.Buckets {
		for i := range report.Instances {
			if strings.Contains(instance.Key, report.Instances[i].Id) {
				report.Instances[i].Cost += instance.Cost.Value
			}
			for volume := range report.Instances[i].IOWrite {
				if volume == instance.Key {
					report.Instances[i].Cost += instance.Cost.Value
				}
			}
			for volume := range report.Instances[i].IORead {
				if volume == instance.Key {
					report.Instances[i].Cost += instance.Cost.Value
				}
			}
		}
	}
	return report
}

// prepareResponseEc2 parses the results from elasticsearch and returns the EC2 usage report
func prepareResponseEc2(ctx context.Context, resEc2 *elastic.SearchResult, user users.User, tx *sql.Tx) (interface{}, error) {
	var response ResponseEc2
	var reports []Report
	err := json.Unmarshal(*resEc2.Aggregations["top_reports"], &response.TopReports)
	if err != nil {
		return nil, err
	}
	for _, account := range response.TopReports.Buckets {
		if len(account.TopReportsHits.Hits.Hits) > 0 {
			report := account.TopReportsHits.Hits.Hits[0].Source
			for i := range report.Instances {
				report.Instances[i].Cost = 0
			}
			resCost, err := makeElasticSearchCostRequest(ctx, user, tx, report.Account, report.ReportDate)
			if err == nil {
				report = addCostToReport(report, resCost)
			}
			reports = append(reports, report)
		}
	}
	return reports, nil
}

// prepareResponseEc2History parses the results from elasticsearch and returns the EC2 usage report
func prepareResponseEc2History(ctx context.Context, resEc2 *elastic.SearchResult) (interface{}, error) {
	var response ResponseEc2
	reports := []Report{}
	err := json.Unmarshal(*resEc2.Aggregations["top_reports"], &response.TopReports)
	if err != nil {
		return nil, err
	}
	for _, account := range response.TopReports.Buckets {
		if len(account.TopReportsHits.Hits.Hits) > 0 {
			reports = append(reports, account.TopReportsHits.Hits.Hits[0].Source)
		}
	}
	return reports, nil
}