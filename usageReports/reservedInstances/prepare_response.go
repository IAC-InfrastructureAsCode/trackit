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
	"encoding/json"

	"github.com/trackit/jsonlog"
	"gopkg.in/olivere/elastic.v5"

	"github.com/trackit/trackit-server/aws/usageReports"
	"github.com/trackit/trackit-server/aws/usageReports/reservedInstances"
)

type (

	// Structure that allow to parse ES response for ReservedInstances Daily reservations
	ResponseReservedInstancesDaily struct {
		Accounts struct {
			Buckets []struct {
				Dates struct {
					Buckets []struct {
						Time         string `json:"key_as_string"`
						Reservations struct {
							Hits struct {
								Hits []struct {
									Reservation reservedInstances.ReservationReport `json:"_source"`
								} `json:"hits"`
							} `json:"hits"`
						} `json:"reservations"`
					} `json:"buckets"`
				} `json:"dates"`
			} `json:"buckets"`
		} `json:"accounts"`
	}

	// ReservationReport has all the information of an ReservedInstances reservation report
	ReservationReport struct {
		utils.ReportBase
		Service     string      `json:"service"`
		Reservation Reservation `json:"reservation"`
	}

	// Reservation contains the information of an ReservedInstances reservation
	Reservation struct {
		reservedInstances.ReservationBase
		Tags map[string]string `json:"tags"`
	}
)

func getReservedInstancesInstanceReportResponse(oldReservation reservedInstances.ReservationReport) ReservationReport {
	tags := make(map[string]string, 0)
	for _, tag := range oldReservation.Reservation.Tags {
		tags[tag.Key] = tag.Value
	}
	newReservation := ReservationReport{
		ReportBase: oldReservation.ReportBase,
		Service:    oldReservation.Service,
		Reservation: Reservation{
			ReservationBase: oldReservation.Reservation.ReservationBase,
			Tags:            tags,
		},
	}
	return newReservation
}

// prepareResponseReservedInstancesDaily parses the results from elasticsearch and returns an array of ReservedInstances daily reservations report
func prepareResponseReservedInstancesDaily(ctx context.Context, resReservedInstances *elastic.SearchResult) ([]ReservationReport, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	var parsedReservedInstances ResponseReservedInstancesDaily
	reservations := make([]ReservationReport, 0)
	err := json.Unmarshal(*resReservedInstances.Aggregations["accounts"], &parsedReservedInstances.Accounts)
	if err != nil {
		logger.Error("Error while unmarshaling ES ReservedInstances response", err)
		return nil, err
	}
	for _, account := range parsedReservedInstances.Accounts.Buckets {
		var lastDate = ""
		for _, date := range account.Dates.Buckets {
			if date.Time > lastDate {
				lastDate = date.Time
			}
		}
		for _, date := range account.Dates.Buckets {
			if date.Time == lastDate {
				for _, reservation := range date.Reservations.Hits.Hits {
					reservations = append(reservations, getReservedInstancesInstanceReportResponse(reservation.Reservation))
				}
			}
		}
	}
	return reservations, nil
}
