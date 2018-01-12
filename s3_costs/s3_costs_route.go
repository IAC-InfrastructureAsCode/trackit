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

package s3_costs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/trackit/jsonlog"
	"github.com/trackit/trackit2/db"
	"github.com/trackit/trackit2/es"
	"github.com/trackit/trackit2/routes"
	"github.com/trackit/trackit2/users"
	"gopkg.in/olivere/elastic.v5"
)

// esQueryParams will store the parsed query params
type esQueryParams struct {
	dateBegin   time.Time
	dateEnd     time.Time
	accountList []uint
}

func init() {
	routes.MethodMuxer{
		http.MethodGet: routes.H(getS3CostData).With(
			db.RequestTransaction{Db: db.Db},
			users.RequireAuthenticatedUser{},
			routes.Documentation{
				Summary:     "get the s3 costs data",
				Description: "Responds with cost data based on the queryparams passed to it",
			},
			routes.QueryArgs{AwsAccountsQueryArg},
			routes.QueryArgs{DateBeginQueryArg},
			routes.QueryArgs{DateEndQueryArg},
		),
	}.H().Register("/s3_costs")
}

var (
	// AwsAccountsQueryArg allows to get the AWS Account IDs in the URL
	// Parameters with routes.QueryArgs. These AWS Account IDs will be a
	// slice of Uint stored in the routes.Arguments map with itself for key.
	AwsAccountsQueryArg = routes.QueryArg{
		Name:        "accounts",
		Type:        routes.QueryArgUintSlice{},
		Description: "The IDs for many AWS account.",
	}

	// DateBeginQueryArg allows to get the iso8601 begin date in the URL
	// Parameters with routes.QueryArgs. This date will be a
	// time.Time stored in the routes.Arguments map with itself for key.
	DateBeginQueryArg = routes.QueryArg{
		Name:        "begin",
		Type:        routes.QueryArgDate{},
		Description: "The begin date.",
	}

	// DateEndQueryArg allows to get the iso8601 begin date in the URL
	// Parameters with routes.QueryArgs. This date will be a
	// time.Time stored in the routes.Arguments map with itself for key.
	DateEndQueryArg = routes.QueryArg{
		Name:        "end",
		Type:        routes.QueryArgDate{},
		Description: "The end date.",
	}
)

// makeElasticSearchStorageRequest prepares and run the request to retrieve storage usage/cost
func makeElasticSearchStorageRequest(ctx context.Context, parsedParams esQueryParams,
	user users.User) (*elastic.SearchResult, error) {
	l := jsonlog.LoggerFromContextOrDefault(ctx)
	index := es.IndexNameForUser(user, "lineitems")
	searchService := GetS3SpaceElasticSearchParams(
		parsedParams.accountList,
		parsedParams.dateBegin,
		parsedParams.dateEnd,
		es.Client,
		index,
	)
	res, err := searchService.Do(ctx)
	if err != nil {
		l.Error("Query execution failed : "+err.Error(), nil)
		return nil, fmt.Errorf("could not execute the ElasticSearch query")
	}
	return res, nil
}

// makeElasticSearchRequestsRequest prepares and run the request to retrieve requests usage/cost
func makeElasticSearchRequestsRequest(ctx context.Context, parsedParams esQueryParams,
	user users.User) (*elastic.SearchResult, error) {
	l := jsonlog.LoggerFromContextOrDefault(ctx)
	index := es.IndexNameForUser(user, "lineitems")
	searchService := GetS3RequestsElasticSearchParams(
		parsedParams.accountList,
		parsedParams.dateBegin,
		parsedParams.dateEnd,
		es.Client,
		index,
	)
	res, err := searchService.Do(ctx)
	if err != nil {
		l.Error("Query execution failed : "+err.Error(), nil)
		return nil, fmt.Errorf("could not execute the ElasticSearch query")
	}
	return res, nil
}

// makeElasticSearchBandwidthRequest prepares and run the request to retrieve bandwidth usage/cost
func makeElasticSearchBandwidthRequest(ctx context.Context, parsedParams esQueryParams,
	user users.User, bwType string) (*elastic.SearchResult, error) {
	l := jsonlog.LoggerFromContextOrDefault(ctx)
	index := es.IndexNameForUser(user, "lineitems")
	searchService := GetS3BandwidthElasticSearchParams(
		parsedParams.accountList,
		parsedParams.dateBegin,
		parsedParams.dateEnd,
		es.Client,
		index,
		bwType,
	)
	res, err := searchService.Do(ctx)
	if err != nil {
		l.Error("Query execution failed : "+err.Error(), nil)
		return nil, fmt.Errorf("could not execute the ElasticSearch query")
	}
	return res, nil
}

// getS3CostData returns the s3 cost data based on the query params, in JSON format.
func getS3CostData(request *http.Request, a routes.Arguments) (int, interface{}) {
	user := a[users.AuthenticatedUser].(users.User)
	parsedParams := esQueryParams{}
	parsedParams.accountList = a[AwsAccountsQueryArg].([]uint)
	parsedParams.dateBegin = a[DateBeginQueryArg].(time.Time)
	parsedParams.dateEnd = a[DateEndQueryArg].(time.Time)
	resStorage, err := makeElasticSearchStorageRequest(request.Context(), parsedParams, user)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	resRequests, err := makeElasticSearchRequestsRequest(request.Context(), parsedParams, user)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	resBandwidthIn, err := makeElasticSearchBandwidthRequest(request.Context(), parsedParams, user, "In")
	if err != nil {
		return http.StatusInternalServerError, err
	}
	resBandwidthOut, err := makeElasticSearchBandwidthRequest(request.Context(), parsedParams, user, "Out")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	res, err := prepareResponse(request.Context(), resStorage, resRequests, resBandwidthIn, resBandwidthOut)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, res
}
