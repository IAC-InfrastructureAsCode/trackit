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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/trackit/jsonlog"

	taws "github.com/trackit/trackit-server/aws"
	"github.com/trackit/trackit-server/aws/usageReports"
	"github.com/trackit/trackit-server/config"
)

// fetchDailySnapshotsList sends in instanceInfoChan the instances fetched from DescribeSnapshots
// and filled by DescribeSnapshots and getSnapshotStats.
func fetchDailySnapshotsList(ctx context.Context, creds *credentials.Credentials, region string, snapshotChan chan Snapshot) error {
	defer close(snapshotChan)
	start, end := utils.GetCurrentCheckedDay()
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	}))
	svc := ec2.New(sess)
	snapshots, err := svc.DescribeSnapshots(nil)
	if err != nil {
		logger.Error("Error when describing snapshots", err.Error())
		return err
	}
	for _, reservation := range snapshots.Reservations {
		for _, snapshot := range reservation.Snapshots {
			stats := getSnapshotStats(ctx, snapshot, sess, start, end)
			costs := make(map[string]float64, 0)
			snapshotChan <- Snapshot{
				SnapshotBase: SnapshotBase{
					Id:         aws.StringValue(snapshot.SnapshotId),
					Region:     aws.StringValue(snapshot.Placement.AvailabilityZone),
					State:      aws.StringValue(snapshot.State.Name),
					Purchasing: getPurchasingOption(snapshot),
					KeyPair:    aws.StringValue(snapshot.KeyName),
					Type:       aws.StringValue(snapshot.SnapshotType),
					Platform:   getPlatformName(aws.StringValue(snapshot.Platform)),
				},
				Tags:  getSnapshotTag(snapshot.Tags),
				Costs: costs,
				Stats: stats,
			}
		}
	}
	return nil
}

// FetchDailySnapshotsStats fetches the stats of the EC2 snapshots of an AwsAccount
// to import them in ElasticSearch. The stats are fetched from the last hour.
// In this way, FetchSnapshotsStats should be called every hour.
func FetchDailySnapshotsStats(ctx context.Context, awsAccount taws.AwsAccount) error {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	logger.Info("Fetching EC2 snapshot stats", map[string]interface{}{"awsAccountId": awsAccount.Id})
	creds, err := taws.GetTemporaryCredentials(awsAccount, MonitorSnapshotStsSessionName)
	if err != nil {
		logger.Error("Error when getting temporary credentials", err.Error())
		return err
	}
	defaultSession := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(config.AwsRegion),
	}))
	now := time.Now().UTC()
	account, err := utils.GetAccountId(ctx, defaultSession)
	if err != nil {
		logger.Error("Error when getting account id", err.Error())
		return err
	}
	regions, err := utils.FetchRegionsList(ctx, defaultSession)
	if err != nil {
		logger.Error("Error when fetching regions list", err.Error())
		return err
	}
	snapshotChans := make([]<-chan Snapshot, 0, len(regions))
	for _, region := range regions {
		snapshotChan := make(chan Snapshot)
		go fetchDailySnapshotsList(ctx, creds, region, snapshotChan)
		snapshotChans = append(snapshotChans, snapshotChan)
	}
	snapshots := make([]SnapshotReport, 0)
	for snapshot := range merge(snapshotChans...) {
		snapshots = append(snapshots, SnapshotReport{
			ReportBase: utils.ReportBase{
				Account:    account,
				ReportDate: now,
				ReportType: "daily",
			},
			Snapshot: snapshot,
		})
	}
	return importSnapshotsToEs(ctx, awsAccount, snapshots)
}
