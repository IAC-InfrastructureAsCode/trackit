//   Copyright 2019 MSolution.IO
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

package mediaconvert

import (
	"context"
	"time"

	"github.com/trackit/jsonlog"

	"github.com/trackit/trackit/es"
)

const TypeMediaConvertReport = "mediaconvert-report"
const IndexPrefixMediaConvertReport = "mediaconvert-reports"
const TemplateNameMediaConvertReport = "mediaconvert-reports"

// put the ElasticSearch index for *-mediaconvert-reports indices at startup.
func init() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	res, err := es.Client.IndexPutTemplate(TemplateNameMediaConvertReport).BodyString(TemplateMediaConvertReport).Do(ctx)
	if err != nil {
		jsonlog.DefaultLogger.Error("Failed to put ES index MediaConvertReport.", err)
	} else {
		jsonlog.DefaultLogger.Info("Put ES index MediaConvertReport.", res)
		ctxCancel()
	}
}

const TemplateMediaConvertReport = `
{
	"template": "*-mediaconvert-reports",
	"version": 2,
	"mappings": {
		"mediaconvert-report": {
			"properties": {
				"account": {
					"type": "keyword"
				},
				"reportDate": {
					"type": "date"
				},
				"reportType": {
					"type": "keyword"
				},
				"job": {
					"properties": {
						"arn": {
							"type": "keyword"
						},
						"region": {
							"type": "keyword"
						},
						"id": {
							"type": "keyword"
						},
						"billingTagsSource": {
							"type": "keyword"
						},
						"createdAt": {
							"type": "date"
						},
						"currentPhase": {
							"type": "keyword"
						},
						"errorCode": {
							"type": "long"
						},
						"ErrorMessage": {
							"type": "keyword"
						},
						"jobPercentComplete": {
							"type": "long"
						},
						"jobTemplate": {
							"type": "keyword"
						},
						"outputGroupDetails": {
							"type": "nested",
							"properties": {
								"outputDetails": {
									"type": "nested",
									"properties": {
										"durationInMs": {
											"type": "long"
										},
										"heightInPx": {
											"type": "long"
										},
										"widthInPx": {
											"type": "long"
										}
									}
								}
							}
						},
						"queue": {
							"type": "keyword"
						},
						"retryCount": {
							"type": "long"
						},
						"role": {
							"type": "keyword"
						},
						"status": {
							"type": "keyword"
						},
						"statusUpdateInterval": {
							"type": "keyword"
						},
						"finishTime": {
							"type": "date"
						},
						"startTime": {
							"type": "date"
						},
						"submitTime": {
							"type": "date"
						},
						"userMetadata": {
							"type": "nested",
							"properties": {
								"key": {
									"type": "keyword"
								},
								"value": {
									"type": "keyword"
								}
							}
						},
						"cost": {
							"type": "double"
						}
					}
				}
			},
			"_all": {
				"enabled": false
			},
			"numeric_detection": false,
			"date_detection": false
		}
	}
}
`
