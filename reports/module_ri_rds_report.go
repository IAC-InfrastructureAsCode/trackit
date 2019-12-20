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

package reports

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/trackit/jsonlog"

	"github.com/trackit/trackit/aws"
	"github.com/trackit/trackit/aws/usageReports/history"
	"github.com/trackit/trackit/usageReports/riRds"
	"github.com/trackit/trackit/users"
)

const riRdsReportSheetName = "Reserved Instances Rds Report"

var riRdsReportModule = module{
	Name:          "Reserved Instances Rds Report",
	SheetName:     riRdsReportSheetName,
	ErrorName:     "riRdsReportError",
	GenerateSheet: generateRiRdsReportSheet,
}

// generateRiRdsReportSheet will generate a sheet with Ri Rds usage report
// It will get data for given AWS account and for a given date
func generateRiRdsReportSheet(ctx context.Context, aas []aws.AwsAccount, date time.Time, tx *sql.Tx, file *excelize.File) (err error) {
	if date.IsZero() {
		date, _ = history.GetHistoryDate()
	}
	return riRdsReportGenerateSheet(ctx, aas, date, tx, file)
}

func riRdsReportGenerateSheet(ctx context.Context, aas []aws.AwsAccount, date time.Time, tx *sql.Tx, file *excelize.File) (err error) {
	data, err := riRdsReportGetData(ctx, aas, date, tx)
	if err == nil {
		return riRdsReportInsertDataInSheet(aas, file, data)
	}
	return
}

func riRdsReportGetData(ctx context.Context, aas []aws.AwsAccount, date time.Time, tx *sql.Tx) (reports []riRds.ReservationReport, err error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	identities := getAwsIdentities(aas)
	user, err := users.GetUserWithId(tx, aas[0].UserId)
	if err != nil {
		return
	}
	parameters := riRds.ReservedInstancesQueryParams{
		AccountList: identities,
		Date:        date,
	}
	logger.Debug("Getting Ri Rds Usage Report for accounts", map[string]interface{}{
		"accounts": aas,
		"date":     date,
	})
	_, reports, err = riRds.GetReservedInstancesData(ctx, parameters, user, tx)
	if err != nil {
		logger.Error("An error occurred while generating an Ri Rds Usage Report", map[string]interface{}{
			"error":    err,
			"accounts": aas,
			"date":     date,
		})
	}
	return
}

func riRdsReportInsertDataInSheet(aas []aws.AwsAccount, file *excelize.File, data []riRds.ReservationReport) (err error) {
	file.NewSheet(riRdsReportSheetName)
	riRdsReportGenerateHeader(file)
	line := 4
	toLine := 0
	for _, report := range data {
		account := getAwsAccount(report.Account, aas)
		formattedAccount := report.Account
		if account != nil {
			formattedAccount = formatAwsAccount(*account)
		}
		instance := report.Reservation
		for currentLine, recurringCharge := range instance.RecurringCharges {
			recurringCells := cells{
				newCell(recurringCharge.Amount, "L"+strconv.Itoa(currentLine + line)).addStyles("price"),
				newCell(recurringCharge.Frequency, "M"+strconv.Itoa(currentLine + line)),
			}
			recurringCells.addStyles("borders", "centerText").setValues(file, riRdsReportSheetName)
			toLine = currentLine + line
		}
		cells := cells{
			newCell(formattedAccount, "A"+strconv.Itoa(line)).mergeTo("A"+strconv.Itoa(toLine)),
			newCell(instance.DBInstanceIdentifier, "B"+strconv.Itoa(line)).mergeTo("B"+strconv.Itoa(toLine)),
			newCell(instance.DBInstanceOfferingId, "C"+strconv.Itoa(line)).mergeTo("C"+strconv.Itoa(toLine)),
			newCell(instance.AvailabilityZone, "D"+strconv.Itoa(line)).mergeTo("D"+strconv.Itoa(toLine)),
			newCell(instance.DBInstanceClass, "E"+strconv.Itoa(line)).mergeTo("E"+strconv.Itoa(toLine)),
			newCell(instance.OfferingType, "F"+strconv.Itoa(line)).mergeTo("F"+strconv.Itoa(toLine)),
			newCell(instance.DBInstanceCount, "G"+strconv.Itoa(line)).mergeTo("G"+strconv.Itoa(toLine)),
			newCell(instance.MultiAZ, "H"+strconv.Itoa(line)).mergeTo("H"+strconv.Itoa(toLine)),
			newCell(instance.State, "I"+strconv.Itoa(line)).mergeTo("I"+strconv.Itoa(toLine)),
			newCell(instance.StartTime.Format("2006-01-02T15:04:05"), "J"+strconv.Itoa(line)).mergeTo("J"+strconv.Itoa(toLine)),
			newCell(instance.EndDate.Format("2006-01-02T15:04:05"), "K"+strconv.Itoa(line)).mergeTo("K"+strconv.Itoa(toLine)),
		}
		cells.addStyles("borders", "centerText").setValues(file, riRdsReportSheetName)
		line++
	}
	return
}

func riRdsReportGenerateHeader(file *excelize.File) {
	header := cells{
		newCell("Account", "A1").mergeTo("A3"),
		newCell("Reservation", "B1").mergeTo("M1"),
		newCell("ID", "B2").mergeTo("B3"),
		newCell("Offering ID", "C2").mergeTo("C3"),
		newCell("Region", "D2").mergeTo("D3"),
		newCell("Class", "E2").mergeTo("E3"),
		newCell("Type", "F2").mergeTo("F3"),
		newCell("Count", "G2").mergeTo("G3"),
		newCell("MultiAZ", "H2").mergeTo("H3"),
		newCell("State", "I2").mergeTo("I3"),
		newCell("Start Date", "J2").mergeTo("J3"),
		newCell("End Date", "K2").mergeTo("K3"),
		newCell("Recurring Charges", "L2").mergeTo("M2"),
		newCell("Amount", "L3"),
		newCell("Frequency", "M3"),
	}
	header.addStyles("borders", "bold", "centerText").setValues(file, riRdsReportSheetName)
	columns := columnsWidth{
		newColumnWidth("A", 30),
		newColumnWidth("B", 30),
		newColumnWidth("C", 35),
		newColumnWidth("D", 15).toColumn("E"),
		newColumnWidth("F", 20),
		newColumnWidth("H", 10),
		newColumnWidth("K", 25),
		newColumnWidth("J", 25),
	}
	columns.setValues(file, riRdsReportSheetName)
	return
}
