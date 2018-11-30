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

package reservedInstances

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/olivere/elastic.v5"

	"github.com/trackit/jsonlog"
	"github.com/trackit/trackit-server/aws/usageReports/reservedInstances"
	"github.com/trackit/trackit-server/errors"
	"github.com/trackit/trackit-server/es"
	"github.com/trackit/trackit-server/users"
)

// makeElasticSearchRequest prepares and run an ES request
// based on the reservedReservationsQueryParams and search params
// It will return the data, an http status code (as int) and an error.
// Because an error can be generated, but is not critical and is not needed to be known by
// the user (e.g if the index does not exists because it was not yet indexed ) the error will
// be returned, but instead of having a 500 status code, it will return the provided status code
// with empty data
func makeElasticSearchRequest(ctx context.Context, parsedParams ReservedInstancesQueryParams,
	esSearchParams func(ReservedInstancesQueryParams, *elastic.Client, string) *elastic.SearchService) (*elastic.SearchResult, int, error) {
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
			return nil, http.StatusOK, errors.GetErrorMessage(ctx, err)
		} else if cast, ok := err.(*elastic.Error); ok && cast.Details.Type == "search_phase_execution_exception" {
			l.Error("Error while getting data from ES", map[string]interface{}{
				"type":  fmt.Sprintf("%T", err),
				"error": err,
			})
		} else {
			l.Error("Query execution failed", map[string]interface{}{"error": err.Error()})
		}
		return nil, http.StatusInternalServerError, errors.GetErrorMessage(ctx, err)
	}
	return res, http.StatusOK, nil
}

// GetReservedInstancesMonthlyReservations does an elastic request and returns an array of reservations monthly report based on query params
func GetReservedInstancesMonthlyReservations(ctx context.Context, params ReservedInstancesQueryParams) (int, []ReservationReport, error) {
	res, returnCode, err := makeElasticSearchRequest(ctx, params, getElasticSearchReservedInstancesMonthlyParams)
	if err != nil {
		return returnCode, nil, err
	}
	reservations, err := prepareResponseReservedInstancesMonthly(ctx, res)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, reservations, nil
}

// GetReservedInstancesDailyReservations does an elastic request and returns an array of reservations daily report based on query params
func GetReservedInstancesDailyReservations(ctx context.Context, params ReservedInstancesQueryParams, user users.User, tx *sql.Tx) (int, []ReservationReport, error) {
	res, returnCode, err := makeElasticSearchRequest(ctx, params, getElasticSearchReservedInstancesDailyParams)
	if err != nil {
		return returnCode, nil, err
	}
	accountsAndIndexes, returnCode, err := es.GetAccountsAndIndexes(params.AccountList, user, tx, es.IndexPrefixLineItems)
	if err != nil {
		return returnCode, nil, err
	}
	params.AccountList = accountsAndIndexes.Accounts
	params.IndexList = accountsAndIndexes.Indexes
	costRes, _, _ := makeElasticSearchRequest(ctx, params, getElasticSearchCostParams)
	reservations, err := prepareResponseReservedInstancesDaily(ctx, res, costRes)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, reservations, nil
}

// GetReservedInstancesData gets ReservedInstances monthly reports based on query params, if there isn't a monthly report, it gets daily reports
func GetReservedInstancesData(ctx context.Context, parsedParams ReservedInstancesQueryParams, user users.User, tx *sql.Tx) (int, []ReservationReport, error) {
	accountsAndIndexes, returnCode, err := es.GetAccountsAndIndexes(parsedParams.AccountList, user, tx, reservedInstances.IndexPrefixReservedInstancesReport)
	if err != nil {
		return returnCode, nil, err
	}
	parsedParams.AccountList = accountsAndIndexes.Accounts
	parsedParams.IndexList = accountsAndIndexes.Indexes
	returnCode, monthlyReservations, err := GetReservedInstancesMonthlyReservations(ctx, parsedParams)
	if err != nil {
		return returnCode, nil, err
	} else if monthlyReservations != nil && len(monthlyReservations) > 0 {
		return returnCode, monthlyReservations, nil
	}
	returnCode, dailyReservations, err := GetReservedInstancesDailyReservations(ctx, parsedParams, user, tx)
	if err != nil {
		return returnCode, nil, err
	}
	return returnCode, dailyReservations, nil
}
