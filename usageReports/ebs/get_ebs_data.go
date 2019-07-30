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

package ebs

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"errors"

	"gopkg.in/olivere/elastic.v5"
	"github.com/trackit/jsonlog"

	"github.com/trackit/trackit-server/aws/usageReports/ebs"
	terrors "github.com/trackit/trackit-server/errors"
	"github.com/trackit/trackit-server/es"
	"github.com/trackit/trackit-server/users"
)

// makeElasticSearchRequest prepares and run an ES request
// based on the ebsQueryParams and search params
// It will return the data, an http status code (as int) and an error.
// Because an error can be generated, but is not critical and is not needed to be known by
// the user (e.g if the index does not exists because it was not yet indexed ) the error will
// be returned, but instead of having a 500 status code, it will return the provided status code
// with empty data
func makeElasticSearchRequest(ctx context.Context, parsedParams EbsQueryParams,
	esSearchParams func(EbsQueryParams, *elastic.Client, string) *elastic.SearchService) (*elastic.SearchResult, int, error) {
	l := jsonlog.LoggerFromContextOrDefault(ctx)
	index := strings.Join(parsedParams.IndexList, ",")
	searchService := esSearchParams(
		parsedParams,
		es.Client,
		index,
	)
	res, err := searchService.Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			l.Warning("Query execution failed, ES index does not exists", map[string]interface{}{
				"index": index,
				"error": err.Error(),
			})
			return nil, http.StatusOK, terrors.GetErrorMessage(ctx, err)
		} else if cast, ok := err.(*elastic.Error); ok && cast.Details.Type == "search_phase_execution_exception" {
			l.Error("Error while getting data from ES", map[string]interface{}{
				"type":  fmt.Sprintf("%T", err),
				"error": err,
			})
		} else {
			l.Error("Query execution failed", map[string]interface{}{"error": err.Error()})
		}
		return nil, http.StatusInternalServerError, terrors.GetErrorMessage(ctx, err)
	}
	return res, http.StatusOK, nil
}

// GetEbsMonthlyInstances does an elastic request and returns an array of instances monthly report based on query params
func GetEbsMonthlyInstances(ctx context.Context, params EbsQueryParams) (int, []InstanceReport, error) {
	res, returnCode, err := makeElasticSearchRequest(ctx, params, getElasticSearchEbsMonthlyParams)
	if err != nil {
		return returnCode, nil, err
	} else if res == nil {
		return http.StatusInternalServerError, nil, errors.New("Error while getting data. Please check again in few hours.")
	}
	instances, err := prepareResponseEbsMonthly(ctx, res)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, instances, nil
}

// GetEbsDailyInstances does an elastic request and returns an array of instances daily report based on query params
func GetEbsDailyInstances(ctx context.Context, params EbsQueryParams, user users.User, tx *sql.Tx) (int, []InstanceReport, error) {
	res, returnCode, err := makeElasticSearchRequest(ctx, params, getElasticSearchEbsDailyParams)
	if err != nil {
		return returnCode, nil, err
	} else if res == nil {
		return http.StatusInternalServerError, nil, errors.New("Error while getting data. Please check again in few hours.")
	}
	accountsAndIndexes, returnCode, err := es.GetAccountsAndIndexes(params.AccountList, user, tx, es.IndexPrefixLineItems)
	if err != nil {
		return returnCode, nil, err
	}
	params.AccountList = accountsAndIndexes.Accounts
	params.IndexList = accountsAndIndexes.Indexes
	costRes, _, _ := makeElasticSearchRequest(ctx, params, getElasticSearchCostParams)
	instances, err := prepareResponseEbsDaily(ctx, res, costRes)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, instances, nil
}

// GetEbsData gets EBS monthly reports based on query params, if there isn't a monthly report, it gets daily reports
func GetEbsData(ctx context.Context, parsedParams EbsQueryParams, user users.User, tx *sql.Tx) (int, []InstanceReport, error) {
	accountsAndIndexes, returnCode, err := es.GetAccountsAndIndexes(parsedParams.AccountList, user, tx, ebs.IndexPrefixEBSReport)
	if err != nil {
		return returnCode, nil, err
	}
	parsedParams.AccountList = accountsAndIndexes.Accounts
	parsedParams.IndexList = accountsAndIndexes.Indexes
	returnCode, monthlyInstances, err := GetEbsMonthlyInstances(ctx, parsedParams)
	if err != nil {
		return returnCode, nil, err
	} else if monthlyInstances != nil && len(monthlyInstances) > 0 {
		return returnCode, monthlyInstances, nil
	}
	returnCode, dailyInstances, err := GetEbsDailyInstances(ctx, parsedParams, user, tx)
	if err != nil {
		return returnCode, nil, err
	}
	return returnCode, dailyInstances, nil
}

// GetEbsUnusedData gets EBS reports and parse them based on query params to have an array of unused instances
func GetEbsUnusedData(ctx context.Context, params EbsUnusedQueryParams, user users.User, tx *sql.Tx) (int, []InstanceReport, error) {
	returnCode, instances, err := GetEbsData(ctx, EbsQueryParams{params.AccountList, nil, params.Date}, user, tx)
	if err != nil {
		return returnCode, nil, err
	}
	return prepareResponseEbsUnused(params, instances)
}
