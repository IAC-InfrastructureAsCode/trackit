//   Copyright 2017-2018 MSolution.IO
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

package tags

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/trackit/jsonlog"
	"gopkg.in/olivere/elastic.v5"

	"github.com/trackit/trackit-server/es"
)

type (
	// structs that allows to parse ES result
	esTagsValuesResult struct {
		Keys struct {
			Buckets []struct {
				Key  string       `json:"key"`
				Tags esTagsResult `json:"tags"`
			} `json:"buckets"`
		} `json:"keys"`
	}

	esTagsResult struct {
		Buckets []struct {
			Tag string `json:"key"`
			Rev struct {
				Filter esFilterResult `json:"filter"`
			} `json:"rev"`
		} `json:"buckets"`
	}

	esFilterResult struct {
		Buckets []struct {
			Product string `json:"key"`
			Cost    struct {
				Value float64 `json:"value"`
			} `json:"cost"`
		} `json:"buckets"`
	}

	// contain a product and his cost
	TagValue struct {
		Item string  `json:"item"`
		Cost float64 `json:"cost"`
	}

	// contain a tag and the list of products associated
	TagsValues struct {
		Tag   string     `json:"tag"`
		Costs []TagValue `json:"costs"`
	}

	// response format of the endpoint
	TagsValuesResponse map[string][]TagsValues
)

// getTagsValuesWithParsedParams will parse the data from ElasticSearch and return it
func getTagsValuesWithParsedParams(ctx context.Context, params tagsValuesQueryParams) (int, interface{}) {
	response := TagsValuesResponse{}
	l := jsonlog.LoggerFromContextOrDefault(ctx)
	var typedDocument esTagsValuesResult
	res, returnCode, err := makeElasticSearchRequestForTagsValues(ctx, params, es.Client)
	if err != nil {
		if returnCode == http.StatusOK {
			return returnCode, response
		}
		return returnCode, errors.New("Internal server error")
	}
	err = json.Unmarshal(*res.Aggregations["data"], &typedDocument)
	if err != nil {
		l.Error("Error while unmarshaling", err)
		return http.StatusInternalServerError, errors.New("Internal server error")
	}
	for _, key := range typedDocument.Keys.Buckets {
		var values []TagsValues
		for _, tag := range key.Tags.Buckets {
			var costs []TagValue
			for _, cost := range tag.Rev.Filter.Buckets {
				costs = append(costs, TagValue{cost.Product, cost.Cost.Value})
			}
			values = append(values, TagsValues{tag.Tag, costs})
		}
		if len(params.TagsKeys) == 0 || arrayContainsString(params.TagsKeys, key.Key) {
			response[key.Key] = values
		}
	}
	return http.StatusOK, response
}

// makeElasticSearchRequestForTagsValues will make the actual request to the ElasticSearch
// It will return the data, an http status code (as int) and an error.
// Because an error can be generated, but is not critical and is not needed to be known by
// the user (e.g if the index does not exists because it was not yet indexed ) the error will
// be returned, but instead of having a 500 status code, it will return the provided status code
// with empty data
func makeElasticSearchRequestForTagsValues(ctx context.Context, params tagsValuesQueryParams, client *elastic.Client) (*elastic.SearchResult, int, error) {
	l := jsonlog.LoggerFromContextOrDefault(ctx)
	filter := getTagsValuesFilter(params.By)
	query := getTagsValuesQuery(params)
	index := strings.Join(params.IndexList, ",")
	search := client.Search().Index(index).Size(0).Query(query)
	search.Aggregation("data", elastic.NewNestedAggregation().Path("tags").
		SubAggregation("keys", elastic.NewTermsAggregation().Field("tags.key").
			SubAggregation("tags", elastic.NewTermsAggregation().Field("tags.tag").
				SubAggregation("rev", elastic.NewReverseNestedAggregation().
					SubAggregation("filter", elastic.NewTermsAggregation().Field(filter).Size(0x7FFFFFFF).
						SubAggregation("cost", elastic.NewSumAggregation().Field("unblendedCost")))))))
	res, err := search.Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			l.Warning("Query execution failed, ES index does not exists", map[string]interface{}{"index": index, "error": err.Error()})
			return nil, http.StatusOK, err
		}
		l.Error("Query execution failed", err.Error())
		return nil, http.StatusInternalServerError, err
	}
	return res, http.StatusOK, nil
}

// getTagsValuesQuery will generate a query for the ElasticSearch based on params
func getTagsValuesQuery(params tagsValuesQueryParams) *elastic.BoolQuery {
	query := elastic.NewBoolQuery()
	if len(params.AccountList) > 0 {
		query = query.Filter(createQueryAccountFilter(params.AccountList))
	}
	query = query.Filter(elastic.NewRangeQuery("usageStartDate").
		From(params.DateBegin).To(params.DateEnd))
	return query
}

// createQueryAccountFilter creates and return a new *elastic.TermsQuery on the accountList array
func createQueryAccountFilter(accountList []string) *elastic.TermsQuery {
	accountListFormatted := make([]interface{}, len(accountList))
	for i, v := range accountList {
		accountListFormatted[i] = v
	}
	return elastic.NewTermsQuery("usageAccountId", accountListFormatted...)
}

// getTagsValuesFilter returns a string of the field to filter
func getTagsValuesFilter(filter string) string {
	var filters = map[string]string{
		"product":          "productCode",
		"region":           "region",
		"account":          "usageAccountId",
		"availabilityzone": "availabilityZone",
	}
	for i := range filters {
		if i == filter {
			return filters[i]
		}
	}
	return "error"
}

// arrayContainsString returns true if a string is present in an array of string
func arrayContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
